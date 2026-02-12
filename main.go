package main

import (
	"fmt"
	"log"

	"free-proxy-list-speed-checker/internal/cache"
	"free-proxy-list-speed-checker/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	c, err := cache.New(cfg.Options.CacheDir)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App name: %s\nSource repo URL: %s\n", cfg.AppName, cfg.SourceRepoUrl)
	fmt.Printf("Cache directory: %s\n", c.Dir)
}
