# RSS aggreGator
Project for [boot.dev](https://www.boot.dev) bootcamp. Gator is a simple CLI tool for aggregating RSS feeds.

# Requirements
 - Go version 1.24 or above.
 - PostgreSQL version 15 or above.

# Installation
Run `go install https://github.com/KMordasewicz/gator`.

## Config

Create a `.gatorconfig.json` file in your home directory with the following structure:

```json
{
  "db_url": "postgres://username:@localhost:5432/database?sslmode=disable"
}
```

Replace the values with your database connection string.

# Usage

## register
 Adds new user and login. Takes username as argument.

 ```
 gator register Bob
 ```

## login
 Change current user. Takes username as argument.
 ```
 gator login Bob
 ```

## users
 Lists all the users.
 ```
 gator users
 ```

## reset
 Factory reset the app.
 ```
 gator reset
 ```

## addfeed
 Saves the feed. Takes name and url as argument.
 ```
 gator addfeed example_feed www.example.com/rss
 ```

## feeds
 Display all the feeds.
```
 gator feeds
 ```

## follow
 Adds a feed to follow for the current user. Takes url as argument.
```
 gator follow www.example.com/rss
 ```

## following
 Display the names of all the feeds followed by current user.
```
 gator following
 ```

## unfollow
 Remove the feed from following for current user. Takes url as argument.
```
 gator unfollow www.example.com/rss
 ```

## agg
 Start aggregatting the feeds in a long-running loop, press CTRL-C to exit. Takes time between fetches as argument e.g. 1s, 1m etc.
```
 gator agg 5m
 ```

## browse
 Display posts from feeds followed by current user. Takes number of feeds to display as argument.
```
 gator browse 10
 ```
