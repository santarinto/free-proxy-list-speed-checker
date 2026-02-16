package cache

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func init() {
	gob.Register("")
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}

type EntryType string

const (
	TypeScalar EntryType = "scalar"
	TypeList   EntryType = "list"
	TypeWeb    EntryType = "web"
)

type Metadata struct {
	Type      EntryType
	CreatedAt time.Time
	UpdatedAt time.Time
}

type RootIndex struct {
	Entries map[string]Metadata
}

type Cache struct {
	Dir       string
	mu        sync.RWMutex
	rootIndex *RootIndex
}

func (c *Cache) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

func (c *Cache) getFilePath(key string) string {
	hashedKey := c.hashKey(key)
	return filepath.Join(c.Dir, hashedKey+".bin")
}

func (c *Cache) getRootIndexPath() string {
	return filepath.Join(c.Dir, "root.index.bin")
}

func (c *Cache) saveToFile(filePath string, data interface{}) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}
	defer file.Close()

	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode data to file %s: %w", filePath, err)
	}

	return nil
}

func (c *Cache) loadFromFile(filePath string, target interface{}) error {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("failed to decode data from file %s: %w", filePath, err)
	}

	return nil
}

func (c *Cache) scanDirectory() error {
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	diskFiles := make(map[string]bool)
	for _, entry := range entries {
		if !entry.IsDir() {
			diskFiles[entry.Name()] = true
		}
	}

	for filename := range diskFiles {
		if filename == "root.index.bin" {
			continue
		}

		found := false
		for key := range c.rootIndex.Entries {
			if c.hashKey(key)+".bin" == filename {
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("warning: orphaned cache file found: %s\n", filename)
		}
	}

	keysToRemove := []string{}
	for key := range c.rootIndex.Entries {
		filePath := c.getFilePath(key)
		if _, exists := diskFiles[filepath.Base(filePath)]; !exists {
			fmt.Printf("warning: cache entry %s referenced in index but file not found: %s\n", key, filePath)
			keysToRemove = append(keysToRemove, key)
		}
	}

	for _, key := range keysToRemove {
		delete(c.rootIndex.Entries, key)
	}

	return nil
}

func (c *Cache) loadRootIndex() error {
	rootIndexPath := c.getRootIndexPath()

	index := &RootIndex{
		Entries: make(map[string]Metadata),
	}

	if err := c.loadFromFile(rootIndexPath, index); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	c.rootIndex = index
	return nil
}

func (c *Cache) saveRootIndex() error {
	rootIndexPath := c.getRootIndexPath()
	return c.saveToFile(rootIndexPath, c.rootIndex)
}

func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)

	wrapper := map[string]interface{}{
		"data": value,
	}

	if err := c.saveToFile(filePath, wrapper); err != nil {
		return err
	}

	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeScalar,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

func (c *Cache) Get(key string) (interface{}, bool, error) {
	c.mu.RLock()
	metadata, exists := c.rootIndex.Entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false, nil
	}

	if metadata.Type != TypeScalar {
		return nil, false, fmt.Errorf("key %s is not a scalar entry", key)
	}

	filePath := c.getFilePath(key)
	wrapper := make(map[string]interface{})
	if err := c.loadFromFile(filePath, &wrapper); err != nil {
		return nil, false, err
	}

	value, exists := wrapper["data"]
	if !exists {
		return nil, false, fmt.Errorf("data field missing in cached entry for key %s", key)
	}

	return value, true, nil
}

func (c *Cache) SetList(key string, items []interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	if err := c.saveToFile(filePath, items); err != nil {
		return err
	}

	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeList,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

func (c *Cache) GetList(key string) ([]interface{}, error) {
	c.mu.RLock()
	metadata, exists := c.rootIndex.Entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("key %s not found in cache", key)
	}

	if metadata.Type != TypeList {
		return nil, fmt.Errorf("key %s is not a list entry", key)
	}

	filePath := c.getFilePath(key)
	var items []interface{}
	if err := c.loadFromFile(filePath, &items); err != nil {
		return nil, err
	}

	return items, nil
}

func (c *Cache) GetWeb(url string) ([]byte, error) {
	key := url

	c.mu.RLock()
	metadata, exists := c.rootIndex.Entries[key]
	c.mu.RUnlock()

	if exists && metadata.Type == TypeWeb {
		filePath := c.getFilePath(key)
		var content []byte
		if err := c.loadFromFile(filePath, &content); err != nil {
			return nil, err
		}
		return content, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download from %s: status code %d", url, resp.StatusCode)
	}

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	if err := c.saveToFile(filePath, content); err != nil {
		return nil, err
	}

	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeWeb,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return content, nil
}

func (c *Cache) Close() error {
	fmt.Println("Saving root index before exit...")
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saveRootIndex()
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
		Dir: dir,
	}

	if err := c.loadRootIndex(); err != nil {
		return nil, fmt.Errorf("failed to load root index: %w", err)
	}

	if err := c.scanDirectory(); err != nil {
		return nil, fmt.Errorf("failed to scan cache directory: %w", err)
	}

	return c, nil
}
