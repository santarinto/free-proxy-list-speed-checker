package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	AppName       string `toml:"app_name"`
	SourceRepoUrl string `toml:"source_repo_url"`
}

func main() {
	configPath := flag.String("config", "config.toml", "Path to config file")
	flag.Parse()

	const configDir = "config"

	resolvedPath := *configPath
	if !filepath.IsAbs(resolvedPath) && filepath.Dir(resolvedPath) == "." {
		resolvedPath = filepath.Join(configDir, resolvedPath)
	}

	resolvedPath = filepath.Clean(resolvedPath)
	if abs, err := filepath.Abs(resolvedPath); err == nil {
		resolvedPath = abs
	}

	if _, err := os.Stat(resolvedPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Config file %s not found", resolvedPath)
		}
		log.Fatalf("Cannot access config file %s: %v", resolvedPath, err)
	}

	var config Config
	if _, err := toml.DecodeFile(resolvedPath, &config); err != nil {
		fmt.Println(err)
	}

	fmt.Printf("App name: %s, Source repo URL: %s\n", config.AppName, config.SourceRepoUrl)
}
