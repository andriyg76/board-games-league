package cache

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

// CacheEntry represents a cached value with metadata
type CacheEntry[K comparable, V any] struct {
	Key       K
	Value     V
	ExpiresAt time.Time
	LastUsed  time.Time
}

// LRUCache is a thread-safe LRU cache with TTL support
type LRUCache[K comparable, V any] struct {
	maxSize    int
	defaultTTL time.Duration
	entries    map[K]*list.Element
	list       *list.List // Doubly linked list for LRU
	mu         sync.RWMutex
}

// NewLRUCache creates a new LRU cache
func NewLRUCache[K comparable, V any](maxSize int, defaultTTL time.Duration) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		maxSize:    maxSize,
		defaultTTL: defaultTTL,
		entries:    make(map[K]*list.Element),
		list:       list.New(),
	}
}

// Get retrieves a value from cache
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.entries[key]
	if !ok {
		var zero V
		return zero, false
	}

	entry := elem.Value.(*CacheEntry[K, V])

	// Check expiration
	if time.Now().After(entry.ExpiresAt) {
		c.removeElement(elem)
		var zero V
		return zero, false
	}

	// Update LRU
	entry.LastUsed = time.Now()
	c.list.MoveToFront(elem)

	return entry.Value, true
}

// Set stores a value in cache
func (c *LRUCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.defaultTTL
	}

	expiresAt := time.Now().Add(ttl)

	// Update existing entry
	if elem, ok := c.entries[key]; ok {
		entry := elem.Value.(*CacheEntry[K, V])
		entry.Value = value
		entry.ExpiresAt = expiresAt
		entry.LastUsed = time.Now()
		c.list.MoveToFront(elem)
		return
	}

	// Add new entry
	entry := &CacheEntry[K, V]{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
		LastUsed:  time.Now(),
	}

	// Evict if needed
	if c.list.Len() >= c.maxSize {
		c.evictLRU()
	}

	elem := c.list.PushFront(entry)
	c.entries[key] = elem
}

// Remove deletes a specific key
func (c *LRUCache[K, V]) Remove(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.entries[key]; ok {
		c.removeElement(elem)
	}
}

// RemoveIf removes entries matching predicate
func (c *LRUCache[K, V]) RemoveIf(predicate func(K, V) bool) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	var toRemove []*list.Element

	for key, elem := range c.entries {
		entry := elem.Value.(*CacheEntry[K, V])
		if predicate(key, entry.Value) {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
		removed++
	}

	return removed
}

// RemoveByKeyFunc removes entries where key matches predicate
func (c *LRUCache[K, V]) RemoveByKeyFunc(predicate func(K) bool) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	var toRemove []*list.Element

	for key, elem := range c.entries {
		if predicate(key) {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
		removed++
	}

	return removed
}

// RemoveByValueFunc removes entries where value matches predicate
func (c *LRUCache[K, V]) RemoveByValueFunc(predicate func(V) bool) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	removed := 0
	var toRemove []*list.Element

	for _, elem := range c.entries {
		entry := elem.Value.(*CacheEntry[K, V])
		if predicate(entry.Value) {
			toRemove = append(toRemove, elem)
		}
	}

	for _, elem := range toRemove {
		c.removeElement(elem)
		removed++
	}

	return removed
}

// CleanExpired removes all expired entries and returns count
func (c *LRUCache[K, V]) CleanExpired() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	removed := 0
	var toRemove []*list.Element

	// Find expired entries
	for elem := c.list.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*CacheEntry[K, V])
		if now.After(entry.ExpiresAt) {
			toRemove = append(toRemove, elem)
		}
	}

	// Remove expired entries
	for _, elem := range toRemove {
		c.removeElement(elem)
		removed++
	}

	return removed
}

// Clear removes all entries
func (c *LRUCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[K]*list.Element)
	c.list = list.New()
}

// Size returns current cache size
func (c *LRUCache[K, V]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// GetStats returns cache statistics
func (c *LRUCache[K, V]) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	expiredCount := 0

	// Count expired entries
	for elem := c.list.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*CacheEntry[K, V])
		if now.After(entry.ExpiresAt) {
			expiredCount++
		}
	}

	return CacheStats{
		CurrentSize:  len(c.entries),
		MaxSize:      c.maxSize,
		ExpiredCount: expiredCount,
		TTL:          formatDuration(c.defaultTTL),
		TTLSeconds:   int64(c.defaultTTL.Seconds()),
	}
}

// evictLRU removes the least recently used entry
func (c *LRUCache[K, V]) evictLRU() {
	if c.list.Len() == 0 {
		return
	}
	back := c.list.Back()
	c.removeElement(back)
}

// removeElement removes an element from cache
func (c *LRUCache[K, V]) removeElement(elem *list.Element) {
	entry := elem.Value.(*CacheEntry[K, V])
	delete(c.entries, entry.Key)
	c.list.Remove(elem)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

