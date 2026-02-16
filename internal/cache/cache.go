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
	// Register common types with gob for proper serialization/deserialization
	gob.Register("")
	gob.Register([]interface{}{})
	gob.Register(map[string]interface{}{})
}

// EntryType defines the type of data stored in cache
type EntryType string

const (
	TypeScalar EntryType = "scalar"
	TypeList   EntryType = "list"
	TypeWeb    EntryType = "web"
)

// Metadata holds information about a cache entry
type Metadata struct {
	Type      EntryType
	CreatedAt time.Time
	UpdatedAt time.Time
}

// RootIndex represents the root of the tree structure
type RootIndex struct {
	Entries map[string]Metadata
}

// Cache implements a two-level tree structure for persistence
type Cache struct {
	Dir       string
	mu        sync.RWMutex
	rootIndex *RootIndex
}

// hashKey computes SHA-256 hash of the key and returns hex-encoded filename
func (c *Cache) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// getFilePath returns the full file path for a hashed key
func (c *Cache) getFilePath(key string) string {
	hashedKey := c.hashKey(key)
	return filepath.Join(c.Dir, hashedKey+".bin")
}

// getRootIndexPath returns the path to the root index file
func (c *Cache) getRootIndexPath() string {
	return filepath.Join(c.Dir, "root.index.bin")
}

// saveToFile writes data to disk using gob encoding
// This is a helper function that does NOT acquire locks
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

// loadFromFile reads data from disk using gob decoding
// This is a helper function that does NOT acquire locks
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

// scanDirectory performs reconciliation between the root index and actual files on disk
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

	// Check for files in disk not in index (orphaned files)
	for filename := range diskFiles {
		if filename == "root.index.bin" {
			continue
		}

		// Check if this file is referenced in the root index
		found := false
		for key := range c.rootIndex.Entries {
			if c.hashKey(key)+".bin" == filename {
				found = true
				break
			}
		}

		if !found {
			// Orphaned file - log but don't delete automatically
			fmt.Printf("warning: orphaned cache file found: %s\n", filename)
		}
	}

	// Check for entries in index but not on disk
	keysToRemove := []string{}
	for key := range c.rootIndex.Entries {
		filePath := c.getFilePath(key)
		if _, exists := diskFiles[filepath.Base(filePath)]; !exists {
			fmt.Printf("warning: cache entry %s referenced in index but file not found: %s\n", key, filePath)
			keysToRemove = append(keysToRemove, key)
		}
	}

	// Remove orphaned entries from index
	for _, key := range keysToRemove {
		delete(c.rootIndex.Entries, key)
	}

	return nil
}

// loadRootIndex loads the root index from disk into memory
func (c *Cache) loadRootIndex() error {
	rootIndexPath := c.getRootIndexPath()

	index := &RootIndex{
		Entries: make(map[string]Metadata),
	}

	// Try to load existing root index
	if err := c.loadFromFile(rootIndexPath, index); err != nil {
		// If file doesn't exist, that's okay - we'll create a new one
		if !os.IsNotExist(err) {
			return err
		}
	}

	c.rootIndex = index
	return nil
}

// saveRootIndex saves the root index to disk
func (c *Cache) saveRootIndex() error {
	rootIndexPath := c.getRootIndexPath()
	return c.saveToFile(rootIndexPath, c.rootIndex)
}

// Set stores a scalar value with the given key
func (c *Cache) Set(key string, value interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)

	// Wrap the value in a map to avoid direct interface{} encoding issues with gob
	wrapper := map[string]interface{}{
		"data": value,
	}

	if err := c.saveToFile(filePath, wrapper); err != nil {
		return err
	}

	// Update root index
	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeScalar,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

// Get retrieves a scalar value by key
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

// SetList stores a list of values with the given key
func (c *Cache) SetList(key string, items []interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	if err := c.saveToFile(filePath, items); err != nil {
		return err
	}

	// Update root index
	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeList,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return nil
}

// GetList retrieves a list by key
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

// GetWeb retrieves web content by URL, downloading and caching if necessary
func (c *Cache) GetWeb(url string) ([]byte, error) {
	// Use URL as the key for web resources
	key := url

	c.mu.RLock()
	metadata, exists := c.rootIndex.Entries[key]
	c.mu.RUnlock()

	// If already cached, load and return
	if exists && metadata.Type == TypeWeb {
		filePath := c.getFilePath(key)
		var content []byte
		if err := c.loadFromFile(filePath, &content); err != nil {
			return nil, err
		}
		return content, nil
	}

	// Download the content
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

	// Store in cache
	c.mu.Lock()
	defer c.mu.Unlock()

	filePath := c.getFilePath(key)
	if err := c.saveToFile(filePath, content); err != nil {
		return nil, err
	}

	// Update root index
	c.rootIndex.Entries[key] = Metadata{
		Type:      TypeWeb,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return content, nil
}

// Close flushes the root index to disk
func (c *Cache) Close() error {
	fmt.Println("Saving root index before exit...")
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saveRootIndex()
}

// New creates a new Cache instance with the specified directory
// It performs initial scan and loads the root index
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

	// Load root index from disk
	if err := c.loadRootIndex(); err != nil {
		return nil, fmt.Errorf("failed to load root index: %w", err)
	}

	// Perform directory scan and reconciliation
	if err := c.scanDirectory(); err != nil {
		return nil, fmt.Errorf("failed to scan cache directory: %w", err)
	}

	return c, nil
}
