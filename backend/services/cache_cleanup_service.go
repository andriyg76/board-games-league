package services

import (
	"context"
	"github.com/andriyg76/bgl/cache"
	"github.com/andriyg76/glog"
	"sync"
	"time"
)

// CacheCleanupService manages periodic cleanup of caches
type CacheCleanupService interface {
	// RegisterCache registers a cache for periodic cleanup
	RegisterCache(name string, c cache.CleanableCache)
	// UnregisterCache removes a cache from cleanup
	UnregisterCache(name string)
	// CleanAll cleans all registered caches
	CleanAll() map[string]int
	// GetAllStats returns statistics for all registered caches
	GetAllStats() []cache.CacheStats
	// Start starts periodic cleanup in background
	Start(ctx context.Context, interval time.Duration)
	// Stop stops the cleanup service
	Stop()
}

type cacheCleanupService struct {
	caches map[string]cache.CleanableCache
	mu     sync.RWMutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewCacheCleanupService creates a new cache cleanup service
func NewCacheCleanupService() CacheCleanupService {
	return &cacheCleanupService{
		caches: make(map[string]cache.CleanableCache),
		stopCh: make(chan struct{}),
	}
}

// RegisterCache registers a cache for periodic cleanup
func (s *cacheCleanupService) RegisterCache(name string, c cache.CleanableCache) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.caches[name] = c
	glog.Info("Registered cache for cleanup: %s (current size: %d)", name, c.Size())
}

// UnregisterCache removes a cache from cleanup
func (s *cacheCleanupService) UnregisterCache(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.caches, name)
	glog.Info("Unregistered cache from cleanup: %s", name)
}

// CleanAll cleans all registered caches and returns statistics
func (s *cacheCleanupService) CleanAll() map[string]int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := make(map[string]int)
	for name, c := range s.caches {
		removed := c.CleanExpired()
		stats[name] = removed
		if removed > 0 {
			glog.Info("Cleaned %d expired entries from cache: %s (remaining: %d)",
				removed, name, c.Size())
		}
	}
	return stats
}

// GetAllStats returns statistics for all registered caches
func (s *cacheCleanupService) GetAllStats() []cache.CacheStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allStats []cache.CacheStats

	for name, c := range s.caches {
		// Check if cache implements GetStats() returning single CacheStats
		if statsProvider, ok := c.(interface {
			GetStats() cache.CacheStats
		}); ok {
			stats := statsProvider.GetStats()
			stats.Name = name
			allStats = append(allStats, stats)
		} else if statsProvider, ok := c.(interface {
			GetStats() []cache.CacheStats
		}); ok {
			// For caches that return multiple stats (like IdAndCode, User)
			statsList := statsProvider.GetStats()
			// Stats already have names set, just append
			allStats = append(allStats, statsList...)
		}
	}

	return allStats
}

// Start starts periodic cleanup in background goroutine
func (s *cacheCleanupService) Start(ctx context.Context, interval time.Duration) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		glog.Info("Cache cleanup service started (interval: %v)", interval)

		for {
			select {
			case <-ticker.C:
				stats := s.CleanAll()
				totalRemoved := 0
				for _, count := range stats {
					totalRemoved += count
				}
				if totalRemoved > 0 {
					glog.Info("Cache cleanup completed: removed %d expired entries total", totalRemoved)
				}
			case <-ctx.Done():
				glog.Info("Cache cleanup service stopping (context cancelled)")
				return
			case <-s.stopCh:
				glog.Info("Cache cleanup service stopping (stop requested)")
				return
			}
		}
	}()
}

// Stop stops the cleanup service
func (s *cacheCleanupService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
	glog.Info("Cache cleanup service stopped")
}

