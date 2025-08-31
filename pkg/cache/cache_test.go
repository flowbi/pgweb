package cache

import (
	"testing"
	"time"
)

func TestCache_BasicOperations(t *testing.T) {
	cache := New(5 * time.Second)
	defer cache.Clear()

	// Test Set and Get
	cache.Set("test_key", "test_value", 0)

	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}
	if value.(string) != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := New(100 * time.Millisecond)
	defer cache.Clear()

	cache.Set("expire_key", "expire_value", 100*time.Millisecond)

	// Should be found immediately
	_, found := cache.Get("expire_key")
	if !found {
		t.Error("Expected to find value immediately after setting")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	_, found = cache.Get("expire_key")
	if found {
		t.Error("Expected value to be expired")
	}
}

func TestCache_GenerateKey(t *testing.T) {
	key1 := GenerateKey("component1", "component2")
	key2 := GenerateKey("component1", "component2")
	key3 := GenerateKey("component1", "different")

	if key1 != key2 {
		t.Error("Same components should generate same key")
	}

	if key1 == key3 {
		t.Error("Different components should generate different keys")
	}
}
