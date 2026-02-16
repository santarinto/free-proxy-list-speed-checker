package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCacheBasics(t *testing.T) {
	// Create a temporary cache directory
	tmpDir := filepath.Join(os.TempDir(), "test-cache-"+t.Name())
	defer os.RemoveAll(tmpDir)

	// Create cache instance
	c, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	defer c.Close()

	// Test Set/Get Scalar with map
	testData := map[string]interface{}{
		"name": "test-value",
		"id":   42,
	}
	if err := c.Set("test-key", testData); err != nil {
		t.Fatalf("Failed to set scalar: %v", err)
	}

	value, exists, err := c.Get("test-key")
	if err != nil {
		t.Fatalf("Failed to get scalar: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist")
	}
	if value == nil {
		t.Fatal("Value should not be nil")
	}

	// Test SetList/GetList
	items := []interface{}{"item1", "item2", "item3"}
	if err := c.SetList("test-list", items); err != nil {
		t.Fatalf("Failed to set list: %v", err)
	}

	retrievedList, err := c.GetList("test-list")
	if err != nil {
		t.Fatalf("Failed to get list: %v", err)
	}
	if len(retrievedList) != 3 {
		t.Fatalf("Expected 3 items, got %d", len(retrievedList))
	}
}

func TestCachePersistence(t *testing.T) {
	// Create a temporary cache directory
	tmpDir := filepath.Join(os.TempDir(), "test-cache-persist-"+t.Name())
	defer os.RemoveAll(tmpDir)

	// Create cache, set data, and close
	{
		c, err := New(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
		}

		persistData := map[string]interface{}{
			"value": "persist-value",
			"time":  time.Now().Unix(),
		}

		if err := c.Set("persist-key", persistData); err != nil {
			t.Fatalf("Failed to set value: %v", err)
		}

		if err := c.Close(); err != nil {
			t.Fatalf("Failed to close cache: %v", err)
		}
	}

	// Reopen cache and verify data is still there
	{
		c, err := New(tmpDir)
		if err != nil {
			t.Fatalf("Failed to reopen cache: %v", err)
		}
		defer c.Close()

		value, exists, err := c.Get("persist-key")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		if !exists {
			t.Fatal("Key should persist after reopening")
		}
		if value == nil {
			t.Fatal("Value should not be nil after reopening")
		}
	}
}
