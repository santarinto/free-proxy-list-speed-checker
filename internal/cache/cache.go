package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Cache struct {
	Dir  string
	mu   sync.RWMutex
	data map[string]interface{}
}

func New(cacheDir string) (*Cache, error) {
	if cacheDir == "" {
		return nil, fmt.Errorf("cache directory cannot be empty")
	}

	var dir string
	clean := filepath.Clean(cacheDir)

	if clean == filepath.Clean(os.TempDir()) {
		prefix := fmt.Sprintf("speed-checker-%s-", time.Now().Format("20060102T150405Z"))
		var err error
		dir, err = os.MkdirTemp(os.TempDir(), prefix)
		if err != nil {
			return nil, fmt.Errorf("failed to create temporary cache directory: %w", err)
		}
	} else {
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
		dir = abs
	}

	c := &Cache{
		Dir:  dir,
		data: make(map[string]interface{}),
	}

	return c, nil
}
