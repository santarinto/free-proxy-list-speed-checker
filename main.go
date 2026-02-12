package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	AppName       string `toml:"app_name"`
	SourceRepoUrl string `toml:"source_repo_url"`
}

func main() {
	configPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	if _, err := os.Stat(*configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Config file %s not found", *configPath)
		}
		log.Fatal(err)
	}

	var config Config
	if _, err := toml.DecodeFile(*configPath, &config); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("App name: %s, Source repo URL: %s\n", config.AppName, config.SourceRepoUrl)
}
