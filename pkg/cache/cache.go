package cache

import (
	"crypto/md5"
	"fmt"
	"reflect"
	"sync"
	"time"
	"unsafe"
)

type item struct {
	value     interface{}
	expiresAt time.Time
	size      int64 // Estimated memory size in bytes
}

type Cache struct {
	items       map[string]*item
	mu          sync.RWMutex
	defaultTTL  time.Duration
	maxItems    int   // Maximum number of items (0 = unlimited)
	maxMemory   int64 // Maximum memory usage in bytes (0 = unlimited)
	currentSize int64 // Current memory usage tracking
}

func New(defaultTTL time.Duration) *Cache {
	c := &Cache{
		items:      make(map[string]*item),
		defaultTTL: defaultTTL,
		maxItems:   500,               // Reasonable default max items
		maxMemory:  100 * 1024 * 1024, // Default 100MB memory limit
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// NewWithMaxItems creates a cache with a specific maximum number of items
func NewWithMaxItems(defaultTTL time.Duration, maxItems int) *Cache {
	c := &Cache{
		items:      make(map[string]*item),
		defaultTTL: defaultTTL,
		maxItems:   maxItems,
		maxMemory:  100 * 1024 * 1024, // Default 100MB memory limit
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// NewWithMemoryLimit creates a cache with a specific memory limit
func NewWithMemoryLimit(defaultTTL time.Duration, maxMemoryMB int) *Cache {
	c := &Cache{
		items:      make(map[string]*item),
		defaultTTL: defaultTTL,
		maxItems:   10000,                            // High item limit when using memory-based limiting
		maxMemory:  int64(maxMemoryMB) * 1024 * 1024, // Convert MB to bytes
	}

	// Start cleanup goroutine
	go c.cleanup()

	return c
}

// NewWithoutCleanup creates a cache without starting the cleanup goroutine
// Useful for testing or when cleanup is managed externally
func NewWithoutCleanup(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]*item),
		defaultTTL: defaultTTL,
		maxItems:   500,               // Default max items
		maxMemory:  100 * 1024 * 1024, // Default 100MB memory limit
	}
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == 0 {
		ttl = c.defaultTTL
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// If maxItems is set and we would exceed it, remove oldest expired items first
	if c.maxItems > 0 && len(c.items) >= c.maxItems {
		c.evictExpired()

		// If we still exceed the limit after cleaning expired items, remove oldest items
		if len(c.items) >= c.maxItems {
			c.evictOldest(len(c.items) - c.maxItems + 1)
		}
	}

	// Estimate the memory size of the value
	itemSize := c.estimateSize(value)

	// If replacing existing item, subtract its size first
	if existingItem, exists := c.items[key]; exists {
		c.currentSize -= existingItem.size
	}

	// Check memory limit first (more important than item count)
	if c.maxMemory > 0 && c.currentSize+itemSize > c.maxMemory {
		c.evictToFitMemory(itemSize)
	}

	newItem := &item{
		value:     value,
		expiresAt: time.Now().Add(ttl),
		size:      itemSize,
	}

	c.items[key] = newItem
	c.currentSize += itemSize
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.expiresAt) {
		// Item expired, will be cleaned up later
		return nil, false
	}

	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, exists := c.items[key]; exists {
		c.currentSize -= item.size
		delete(c.items, key)
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*item)
	c.currentSize = 0
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.expiresAt) {
				delete(c.items, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *Cache) Stats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	expired := 0
	now := time.Now()
	for _, item := range c.items {
		if now.After(item.expiresAt) {
			expired++
		}
	}

	return map[string]interface{}{
		"total_items":       len(c.items),
		"expired_items":     expired,
		"active_items":      len(c.items) - expired,
		"memory_used_mb":    c.currentSize / (1024 * 1024),
		"memory_limit_mb":   c.maxMemory / (1024 * 1024),
		"memory_used_bytes": c.currentSize,
	}
}

// estimateSize estimates the memory size of a value in bytes
func (c *Cache) estimateSize(value interface{}) int64 {
	if value == nil {
		return 8 // pointer size
	}

	// Use reflection to get the approximate size
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return int64(len(v.String()) + 16) // string header + data
	case reflect.Slice, reflect.Array:
		size := int64(v.Len()) * 8 // estimate 8 bytes per element as baseline
		// For slice of interfaces or complex types, add more
		if v.Len() > 0 && v.Index(0).Kind() == reflect.Interface {
			size *= 4 // interfaces are more expensive
		}
		return size + 24 // slice header
	case reflect.Map:
		return int64(v.Len()) * 32 // estimate 32 bytes per map entry
	case reflect.Ptr:
		if v.IsNil() {
			return 8
		}
		return 8 + c.estimateSize(v.Elem().Interface())
	case reflect.Struct:
		// For structs, estimate based on number of fields * average field size
		return int64(v.NumField()) * 16
	default:
		return int64(unsafe.Sizeof(value))
	}
}

// evictToFitMemory removes items until there's enough space for newItemSize
func (c *Cache) evictToFitMemory(newItemSize int64) {
	targetSize := c.maxMemory - newItemSize

	// First try to remove expired items
	c.evictExpired()

	// If still not enough space, remove oldest items by size
	if c.currentSize > targetSize {
		c.evictOldestBySize(c.currentSize - targetSize)
	}
}

// evictExpired removes all expired items (called with lock held)
func (c *Cache) evictExpired() {
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiresAt) {
			c.currentSize -= item.size
			delete(c.items, key)
		}
	}
}

// evictOldest removes the N oldest items by expiration time (called with lock held)
func (c *Cache) evictOldest(count int) {
	if count <= 0 {
		return
	}

	// Collect items with their keys and sort by expiration time
	type keyItem struct {
		key  string
		item *item
	}

	items := make([]keyItem, 0, len(c.items))
	for key, item := range c.items {
		items = append(items, keyItem{key, item})
	}

	// Sort by expiration time (oldest first)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].item.expiresAt.After(items[j].item.expiresAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Remove the oldest count items
	for i := 0; i < count && i < len(items); i++ {
		c.currentSize -= items[i].item.size
		delete(c.items, items[i].key)
	}
}

// evictOldestBySize removes items until the specified amount of memory is freed
func (c *Cache) evictOldestBySize(targetBytesToFree int64) {
	if targetBytesToFree <= 0 {
		return
	}

	// Collect items with their keys and sort by expiration time (oldest first)
	type keyItem struct {
		key  string
		item *item
	}

	items := make([]keyItem, 0, len(c.items))
	for key, item := range c.items {
		items = append(items, keyItem{key, item})
	}

	// Sort by expiration time (oldest first)
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].item.expiresAt.After(items[j].item.expiresAt) {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Remove items until we've freed enough memory
	freedBytes := int64(0)
	for i := 0; i < len(items) && freedBytes < targetBytesToFree; i++ {
		freedBytes += items[i].item.size
		c.currentSize -= items[i].item.size
		delete(c.items, items[i].key)
	}
}

// GenerateKey creates a cache key from multiple string components
func GenerateKey(components ...string) string {
	combined := ""
	for _, comp := range components {
		combined += comp + "|"
	}

	hash := md5.Sum([]byte(combined))
	return fmt.Sprintf("%x", hash)
}
