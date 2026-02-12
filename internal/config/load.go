package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

func Load() (*Config, error) {
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

	ext := filepath.Ext(resolvedPath)
	localPath := strings.TrimSuffix(resolvedPath, ext) + ".local" + ext

	if _, err := os.Stat(localPath); err == nil {
		var patch ConfigPath
		if _, err := toml.DecodeFile(localPath, &patch); err != nil {
			fmt.Println(err)
		}
		config.ApplyPatch(patch)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("cannot access local config file %s: %w", localPath, err)
	}

	return &config, nil
}
