package cache

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestCacheBasics(t *testing.T) {
	tmpDir := t.TempDir()

	// Create cache instance
	c, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("cache.Close: %v", err)
		}
	})

	// Test Set/Get Scalar with map
	testData := map[string]interface{}{
		"name": "test-value",
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
	got, ok := value.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", value)
	}
	if got["name"] != "test-value" {
		t.Errorf("Expected name=test-value, got %v", got["name"])
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
	if len(retrievedList) != len(items) {
		t.Fatalf("Expected %d items, got %d", len(items), len(retrievedList))
	}
	for i, expected := range items {
		if retrievedList[i] != expected {
			t.Errorf("Item %d: expected %v, got %v", i, expected, retrievedList[i])
		}
	}
}

func TestCachePersistence(t *testing.T) {
	tmpDir := t.TempDir()

	persistData := map[string]interface{}{
		"value": "persist-value",
		"time":  time.Now().Unix(),
	}

	// Create cache, set data, and close
	{
		c, err := New(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create cache: %v", err)
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
		t.Cleanup(func() {
			if err := c.Close(); err != nil {
				t.Errorf("cache.Close: %v", err)
			}
		})

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
		got, ok := value.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map[string]interface{}, got %T", value)
		}
		if got["value"] != persistData["value"] {
			t.Errorf("Expected value=%v, got %v", persistData["value"], got["value"])
		}
		if got["time"] != persistData["time"] {
			t.Errorf("Expected time=%v, got %v", persistData["time"], got["time"])
		}
	}

	// List persistence: set a list, close, reopen, verify elements survive
	persistList := []interface{}{"alpha", "beta", "gamma"}
	{
		c, err := New(tmpDir)
		if err != nil {
			t.Fatalf("Failed to create cache for list persistence: %v", err)
		}

		if err := c.SetList("persist-list", persistList); err != nil {
			t.Fatalf("Failed to set list: %v", err)
		}

		if err := c.Close(); err != nil {
			t.Fatalf("Failed to close cache after SetList: %v", err)
		}
	}

	{
		c, err := New(tmpDir)
		if err != nil {
			t.Fatalf("Failed to reopen cache for list persistence: %v", err)
		}
		t.Cleanup(func() {
			if err := c.Close(); err != nil {
				t.Errorf("cache.Close: %v", err)
			}
		})

		retrieved, err := c.GetList("persist-list")
		if err != nil {
			t.Fatalf("GetList after reopen: %v", err)
		}
		if len(retrieved) != len(persistList) {
			t.Fatalf("Expected %d items, got %d", len(persistList), len(retrieved))
		}
		for i, expected := range persistList {
			if retrieved[i] != expected {
				t.Errorf("Item %d: expected %v, got %v", i, expected, retrieved[i])
			}
		}
	}
}

func TestCacheErrorPaths(t *testing.T) {
	tmpDir := t.TempDir()

	c, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("cache.Close: %v", err)
		}
	})

	// Get on non-existent key returns (nil, false, nil)
	value, exists, err := c.Get("missing-key")
	if err != nil {
		t.Fatalf("Get on missing key should not error, got: %v", err)
	}
	if exists {
		t.Fatal("Get on missing key: exists should be false")
	}
	if value != nil {
		t.Fatal("Get on missing key: value should be nil")
	}

	// GetList on non-existent key returns ErrNotFound
	_, err = c.GetList("missing-list")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetList on missing key: expected ErrNotFound, got %v", err)
	}

	// Type mismatch: set scalar, then try GetList
	if err := c.Set("scalar-key", map[string]interface{}{"x": 1}); err != nil {
		t.Fatalf("Failed to set scalar: %v", err)
	}
	_, err = c.GetList("scalar-key")
	if err == nil {
		t.Fatal("GetList on scalar key: expected error, got nil")
	}
}

func TestCacheGetWeb(t *testing.T) {
	var requestCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		fmt.Fprint(w, "proxy-list-content")
	}))
	t.Cleanup(srv.Close)

	tmpDir := t.TempDir()
	c, err := New(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("cache.Close: %v", err)
		}
	})

	// First call — cache miss: should fetch from server and persist.
	content, err := c.GetWeb(srv.URL)
	if err != nil {
		t.Fatalf("GetWeb (cache miss): %v", err)
	}
	if string(content) != "proxy-list-content" {
		t.Errorf("GetWeb (cache miss): expected %q, got %q", "proxy-list-content", string(content))
	}
	if requestCount.Load() != 1 {
		t.Errorf("Expected 1 HTTP request after cache miss, got %d", requestCount.Load())
	}

	// Second call — cache hit: should return persisted content without hitting the server.
	content2, err := c.GetWeb(srv.URL)
	if err != nil {
		t.Fatalf("GetWeb (cache hit): %v", err)
	}
	if string(content2) != "proxy-list-content" {
		t.Errorf("GetWeb (cache hit): expected %q, got %q", "proxy-list-content", string(content2))
	}
	if requestCount.Load() != 1 {
		t.Errorf("Expected still 1 HTTP request after cache hit, got %d", requestCount.Load())
	}
}
