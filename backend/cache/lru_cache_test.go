package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewLRUCache(t *testing.T) {
	cache := NewLRUCache[string, int](100, time.Hour)
	assert.NotNil(t, cache)
	assert.Equal(t, 100, cache.maxSize)
	assert.Equal(t, time.Hour, cache.defaultTTL)
	assert.Equal(t, 0, cache.Size())
}

func TestLRUCache_Get_Set(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	// Test Get on empty cache
	value, ok := cache.Get("key1")
	assert.False(t, ok)
	assert.Empty(t, value)

	// Test Set and Get
	cache.Set("key1", "value1", 0)
	value, ok = cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// Test Set with custom TTL
	cache.Set("key2", "value2", 30*time.Minute)
	value, ok = cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestLRUCache_UpdateExisting(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	cache.Set("key1", "value1", 0)
	cache.Set("key1", "value1_updated", 0)

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1_updated", value)
	assert.Equal(t, 1, cache.Size())
}

func TestLRUCache_Expiration(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	// Set with very short TTL
	cache.Set("key1", "value1", 10*time.Millisecond)
	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, "value1", value)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)
	value, ok = cache.Get("key1")
	assert.False(t, ok)
	assert.Empty(t, value)
	assert.Equal(t, 0, cache.Size())
}

func TestLRUCache_Eviction(t *testing.T) {
	cache := NewLRUCache[string, int](3, time.Hour)

	// Fill cache to max size
	cache.Set("key1", 1, 0)
	cache.Set("key2", 2, 0)
	cache.Set("key3", 3, 0)
	assert.Equal(t, 3, cache.Size())

	// Add one more - should evict LRU (key1)
	cache.Set("key4", 4, 0)
	assert.Equal(t, 3, cache.Size())

	// key1 should be evicted
	_, ok := cache.Get("key1")
	assert.False(t, ok)

	// Other keys should still be there
	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 2, value)

	value, ok = cache.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, 3, value)

	value, ok = cache.Get("key4")
	assert.True(t, ok)
	assert.Equal(t, 4, value)
}

func TestLRUCache_LRUOrder(t *testing.T) {
	cache := NewLRUCache[string, int](3, time.Hour)

	cache.Set("key1", 1, 0)
	cache.Set("key2", 2, 0)
	cache.Set("key3", 3, 0)

	// Access key1 to make it most recently used
	cache.Get("key1")

	// Add key4 - should evict key2 (least recently used)
	cache.Set("key4", 4, 0)

	_, ok := cache.Get("key2")
	assert.False(t, ok, "key2 should be evicted")

	// key1, key3, key4 should still be there
	_, ok1 := cache.Get("key1")
	assert.True(t, ok1)
	_, ok3 := cache.Get("key3")
	assert.True(t, ok3)
	_, ok4 := cache.Get("key4")
	assert.True(t, ok4)
}

func TestLRUCache_Remove(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	assert.Equal(t, 2, cache.Size())

	cache.Remove("key1")
	assert.Equal(t, 1, cache.Size())

	_, ok := cache.Get("key1")
	assert.False(t, ok)

	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)
}

func TestLRUCache_RemoveIf(t *testing.T) {
	cache := NewLRUCache[string, int](10, time.Hour)

	cache.Set("key1", 1, 0)
	cache.Set("key2", 2, 0)
	cache.Set("key3", 3, 0)
	cache.Set("key4", 4, 0)

	// Remove entries with value > 2
	removed := cache.RemoveIf(func(key string, value int) bool {
		return value > 2
	})

	assert.Equal(t, 2, removed)
	assert.Equal(t, 2, cache.Size())

	_, ok := cache.Get("key3")
	assert.False(t, ok)
	_, ok = cache.Get("key4")
	assert.False(t, ok)

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)

	value, ok = cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, 2, value)
}

func TestLRUCache_RemoveByKeyFunc(t *testing.T) {
	cache := NewLRUCache[string, int](10, time.Hour)

	cache.Set("key1", 1, 0)
	cache.Set("key2", 2, 0)
	cache.Set("key3", 3, 0)
	cache.Set("key4", 4, 0)

	// Remove keys starting with "key3"
	removed := cache.RemoveByKeyFunc(func(key string) bool {
		return key == "key3" || key == "key4"
	})

	assert.Equal(t, 2, removed)
	assert.Equal(t, 2, cache.Size())

	_, ok := cache.Get("key3")
	assert.False(t, ok)
	_, ok = cache.Get("key4")
	assert.False(t, ok)

	value, ok := cache.Get("key1")
	assert.True(t, ok)
	assert.Equal(t, 1, value)
}

func TestLRUCache_RemoveByValueFunc(t *testing.T) {
	type testStruct struct {
		ID   int
		Name string
	}

	cache := NewLRUCache[string, testStruct](10, time.Hour)

	cache.Set("key1", testStruct{ID: 1, Name: "Alice"}, 0)
	cache.Set("key2", testStruct{ID: 2, Name: "Bob"}, 0)
	cache.Set("key3", testStruct{ID: 3, Name: "Alice"}, 0)

	// Remove entries with Name == "Alice"
	removed := cache.RemoveByValueFunc(func(value testStruct) bool {
		return value.Name == "Alice"
	})

	assert.Equal(t, 2, removed)
	assert.Equal(t, 1, cache.Size())

	_, ok := cache.Get("key1")
	assert.False(t, ok)
	_, ok = cache.Get("key3")
	assert.False(t, ok)

	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "Bob", value.Name)
}

func TestLRUCache_CleanExpired(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	// Set entries with different TTLs
	cache.Set("key1", "value1", 10*time.Millisecond)
	cache.Set("key2", "value2", 50*time.Millisecond)
	cache.Set("key3", "value3", time.Hour)

	time.Sleep(20 * time.Millisecond)

	// Clean expired entries
	removed := cache.CleanExpired()
	assert.Equal(t, 1, removed)
	assert.Equal(t, 2, cache.Size())

	_, ok := cache.Get("key1")
	assert.False(t, ok)

	value, ok := cache.Get("key2")
	assert.True(t, ok)
	assert.Equal(t, "value2", value)

	value, ok = cache.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, "value3", value)
}

func TestLRUCache_Clear(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)
	cache.Set("key3", "value3", 0)
	assert.Equal(t, 3, cache.Size())

	cache.Clear()
	assert.Equal(t, 0, cache.Size())

	_, ok := cache.Get("key1")
	assert.False(t, ok)
}

func TestLRUCache_Size(t *testing.T) {
	cache := NewLRUCache[string, string](10, time.Hour)

	assert.Equal(t, 0, cache.Size())

	cache.Set("key1", "value1", 0)
	assert.Equal(t, 1, cache.Size())

	cache.Set("key2", "value2", 0)
	assert.Equal(t, 2, cache.Size())

	cache.Remove("key1")
	assert.Equal(t, 1, cache.Size())
}

func TestLRUCache_GetStats(t *testing.T) {
	cache := NewLRUCache[string, string](100, 30*time.Minute)

	cache.Set("key1", "value1", 0)
	cache.Set("key2", "value2", 0)

	stats := cache.GetStats()
	assert.Equal(t, 2, stats.CurrentSize)
	assert.Equal(t, 100, stats.MaxSize)
	assert.Equal(t, 0, stats.ExpiredCount)
	assert.Equal(t, int64(1800), stats.TTLSeconds) // 30 minutes in seconds
	assert.NotEmpty(t, stats.TTL)
}

func TestLRUCache_GetStats_WithExpired(t *testing.T) {
	cache := NewLRUCache[string, string](100, time.Hour)

	cache.Set("key1", "value1", 10*time.Millisecond)
	cache.Set("key2", "value2", time.Hour)

	time.Sleep(20 * time.Millisecond)

	stats := cache.GetStats()
	assert.Equal(t, 2, stats.CurrentSize)  // Still in cache until cleaned
	assert.Equal(t, 1, stats.ExpiredCount) // One expired entry
}

func TestLRUCache_ConcurrentAccess(t *testing.T) {
	cache := NewLRUCache[int, int](100, time.Hour)

	// Concurrent writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				cache.Set(id*10+j, id*10+j, 0)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 10; j++ {
				cache.Get(id*10 + j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Cache should still be functional
	assert.True(t, cache.Size() > 0)
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "seconds only",
			duration: 30 * time.Second,
			expected: "30s",
		},
		{
			name:     "minutes and seconds",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5m 30s",
		},
		{
			name:     "hours, minutes and seconds",
			duration: 2*time.Hour + 30*time.Minute + 15*time.Second,
			expected: "2h 30m 15s",
		},
		{
			name:     "days, hours, minutes and seconds",
			duration: 3*24*time.Hour + 5*time.Hour + 30*time.Minute + 15*time.Second,
			expected: "3d 5h 30m 15s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}
