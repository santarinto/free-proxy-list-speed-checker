package cache

import (
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
	defer c.Close()

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
}
