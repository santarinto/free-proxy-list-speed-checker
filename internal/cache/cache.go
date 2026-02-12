package cache

import (
	"fmt"
	"os"
	"path/filepath"
)

type Cache struct {
	Dir string
}

func New(cacheDir string) (*Cache, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory cannot be empty")
	}

	clean := filepath.Clean(cacheDir)

	abs, err := filepath.Abs(clean)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path of cache directory: %w", err)
	}

	if err := os.MkdirAll(abs, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("failed to stat cache directory: %w", err)
	}

	if !info.IsDir() {
		return nil, fmt.Errorf("cache directory %s is not a directory", abs)
	}

	return &Cache{abs}, nil
}
