package cache

// CacheStats represents statistics about a cache
type CacheStats struct {
	Name         string `json:"name"`
	CurrentSize  int    `json:"current_size"`
	MaxSize      int    `json:"max_size"`
	ExpiredCount int    `json:"expired_count"`
	TTL          string `json:"ttl"` // Human-readable TTL
	TTLSeconds   int64  `json:"ttl_seconds"`
}

// CacheStatsProvider provides statistics about cache
type CacheStatsProvider interface {
	GetStats() CacheStats
}

// CleanableCache represents a cache that can be cleaned of expired entries
type CleanableCache interface {
	// CleanExpired removes all expired entries and returns count of removed items
	CleanExpired() int
	// Size returns current number of entries
	Size() int
}

