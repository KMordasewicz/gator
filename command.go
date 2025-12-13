package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/KMordasewicz/blog-aggregator/internal/config"
	"github.com/KMordasewicz/blog-aggregator/internal/database"
	"github.com/KMordasewicz/blog-aggregator/internal/feed"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	callable map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	err := c.callable[cmd.name](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.callable[name] = f
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing username argument")
	}
	name := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, name)
	if err != nil {
		log.Fatalf("User %s doesn't exists\n", name)
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("Username has been successfully set to %s\n", name)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing username argument")
	}
	name := cmd.args[0]
	ctx := context.Background()
	_, err := s.db.GetUser(ctx, name)
	if err == nil {
		log.Fatal("User already exists")
	}
	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
	})
	if err != nil {
		return err
	}
	err = s.cfg.SetUser(name)
	if err != nil {
		return err
	}
	fmt.Printf("User %s was created: %v\n", name, user)
	return nil
}

func handleReset(s *state, cmd command) error {
	ctx := context.Background()
	err := s.db.DeleteUsers(ctx)
	if err != nil {
		log.Fatalf("error while trying to truncate users:\n%v\n", err)
	}
	fmt.Println("Users table was successfully truncated")
	return nil
}

func handleUsers(s *state, cmd command) error {
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		log.Fatalf("Unable to fetch users: %v", err)
	}
	for _, user := range users {
		text := user.Name
		if text == s.cfg.CurrentUserName {
			text += " (current)"
		}
		fmt.Println(text)
	}
	return nil
}

func scrapeFeeds(s *state) {
	ctx := context.Background()
	next_feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		log.Printf("Couldn't get next feed: %v\n", err)
	}
	_, err = s.db.MarkFeedFetched(ctx, next_feed.ID)
	if err != nil {
		log.Printf("Couldn't mark feed as fetched: %v\n", err)
	}
	feeds, err := feed.FetchFeed(ctx, next_feed.Url)
	if err != nil {
		log.Printf("Couldn't fetch feeds: %v\n", err)
	}
	feeds.Clear()
	for _, feed := range feeds.Channel.Item {
		publishDate, err := time.Parse(time.RFC1123Z, feed.PubDate)
		if err != nil {
			log.Printf("Couldn't extract publish date: %v\n", err)
			continue
		}
		post, err := s.db.CreatePost(ctx, database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       feed.Title,
			Url:         feed.Link,
			Description: feed.Description,
			PublishedAt: publishDate,
			FeedID:      next_feed.ID,
		})
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23505" {
					continue
				} else {
					log.Fatal(pqErr)
				}
			} else {
				log.Fatalf("Couldn't add the post to db: %v\n", err)
			}
		}
		fmt.Printf("Successfully saved a post: %v\n", post)
	}
}

func hanldeAgg(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing time_between_reqs argument")
	}
	timeBetweenTicks, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Collecting feeds every %v", timeBetweenTicks)
	ticker := time.NewTicker(timeBetweenTicks)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
	return nil
}

func handleAddFeed(s *state, cmd command, user database.User) error {
	if l := len(cmd.args); l < 2 {
		log.Fatalf("Expected 2 arguments, got %d instead: %v\n", l, cmd.args)
	}
	name := cmd.args[0]
	url := cmd.args[1]
	ctx := context.Background()
	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	})
	if err != nil {
		log.Fatalf("Couldn't add feed: %v\n", err)
	}
	fmt.Printf("Successfully added new feed: %v\n", feed)
	follow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		log.Fatalf("Couldn't add feed to follow: %v", err)
	}
	fmt.Printf("Successfully added new feed follow: %v\n", follow)
	return nil
}

func handleFeeds(s *state, cmd command) error {
	ctx := context.Background()
	feeds, err := s.db.GetAllFeeds(ctx)
	if err != nil {
		log.Fatalf("Couldn't get the feeds: %v\n", err)
	}
	for _, feed := range feeds {
		fmt.Printf("%v\n", feed)
	}
	return nil
}

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing url argument")
	}
	ctx := context.Background()
	url := cmd.args[0]
	feed, err := s.db.GetURLFeeds(ctx, url)
	if err != nil {
		log.Fatalf("Couldn't get feed for %s\n", url)
	}
	follow, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		log.Fatalf("Couldn't create a follow: %v\n", err)
	}
	fmt.Printf("Created follow for %s feed for %s user: %v\n", feed.Name, user.Name, follow)
	return nil
}

func handleFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()
	userName := user.Name
	follows, err := s.db.GetFeedFollowsForUser(ctx, userName)
	if err != nil {
		log.Fatalf("Couldn't get follows for %s user\n", userName)
	}
	for _, follow := range follows {
		fmt.Printf("%v\n", follow.FeedName)
	}
	return nil
}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return errors.New("error: missing url argument")
	}
	url := cmd.args[0]
	ctx := context.Background()
	feed, err := s.db.GetURLFeeds(ctx, url)
	if err != nil {
		return err
	}
	deletedFeedFollow, err := s.db.DeleteFeedFollowByUserAndFeedID(ctx, database.DeleteFeedFollowByUserAndFeedIDParams{UserID: user.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}
	fmt.Printf("Followed feed successfully deleted: %v\n", deletedFeedFollow)
	return nil
}

func handleBrowse(s *state, cmd command, user database.User) error {
	var limit int32 = 2
	if len(cmd.args) > 0 {
		newLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return err
		}
		limit = int32(newLimit)
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		Name:  user.Name,
		Limit: limit,
	})
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("%v\n", post)
	}
	return nil
}
