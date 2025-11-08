package main

import (
	"fmt"
	"log"

	"github.com/KMordasewicz/blog-aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Couldn't read config: %e", err)
	}
	err = cfg.SetUser("Krzysztof")
	if err != nil {
		log.Fatalf("Couldn't set user: %e", err)
	}
	newCfg, err := config.Read()
	if err != nil {
		log.Fatalf("Couldn't read config: %e", err)
	}
	fmt.Printf("%v", newCfg)
}
