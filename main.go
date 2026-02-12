package main

import (
	"fmt"
	"log"

	appconfig "free-proxy-list-speed-checker/internal/config"
)

func main() {
	cfg, err := appconfig.Load()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("App name: %s,\nSource repo URL: %s\n", cfg.AppName, cfg.SourceRepoUrl)
}
