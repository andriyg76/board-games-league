package services

import (
	"time"

	"github.com/andriyg76/bgl/cache"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	IdAndCodeCacheSize = 2000
	IdAndCodeCacheTTL  = 1 * time.Hour
)

type IdAndCodeCache interface {
	GetByID(id primitive.ObjectID) *models.IdAndCode
	GetByCode(code string) (*models.IdAndCode, error)
}

type idAndCodeCacheImpl struct {
	byID   *cache.LRUCache[primitive.ObjectID, *models.IdAndCode]
	byCode *cache.LRUCache[string, *models.IdAndCode]
}

func NewIdAndCodeCache() IdAndCodeCache {
	return &idAndCodeCacheImpl{
		byID:   cache.NewLRUCache[primitive.ObjectID, *models.IdAndCode](IdAndCodeCacheSize, IdAndCodeCacheTTL),
		byCode: cache.NewLRUCache[string, *models.IdAndCode](IdAndCodeCacheSize, IdAndCodeCacheTTL),
	}
}

func (c *idAndCodeCacheImpl) GetByID(id primitive.ObjectID) *models.IdAndCode {
	if idCode, ok := c.byID.Get(id); ok {
		return idCode
	}
	// Create and store
	return c.StoreByID(id)
}

func (c *idAndCodeCacheImpl) GetByCode(code string) (*models.IdAndCode, error) {
	if idCode, ok := c.byCode.Get(code); ok {
		return idCode, nil
	}
	// Create from code and store
	return c.StoreByCode(code)
}

func (c *idAndCodeCacheImpl) Store(idCode *models.IdAndCode) {
	c.byID.Set(idCode.ID, idCode, IdAndCodeCacheTTL)
	c.byCode.Set(idCode.Code, idCode, IdAndCodeCacheTTL)
}

func (c *idAndCodeCacheImpl) StoreByID(id primitive.ObjectID) *models.IdAndCode {
	idCode := models.NewIdAndCode(id)
	c.Store(idCode)
	return idCode
}

func (c *idAndCodeCacheImpl) StoreByCode(code string) (*models.IdAndCode, error) {
	idCode, err := models.NewIdAndCodeFromCode(code)
	if err != nil {
		return nil, err
	}
	c.Store(idCode)
	return idCode, nil
}

func (c *idAndCodeCacheImpl) Remove(key primitive.ObjectID) {
	c.byID.Remove(key)
	// Also remove by code if we have it
	if idCode, ok := c.byID.Get(key); ok {
		c.byCode.Remove(idCode.Code)
	}
}

func (c *idAndCodeCacheImpl) RemoveByKeyFunc(predicate func(primitive.ObjectID) bool) int {
	return c.byID.RemoveByKeyFunc(predicate)
}

func (c *idAndCodeCacheImpl) RemoveByValueFunc(predicate func(*models.IdAndCode) bool) int {
	return c.byID.RemoveByValueFunc(predicate)
}

func (c *idAndCodeCacheImpl) Clear() {
	c.byID.Clear()
	c.byCode.Clear()
}

// CleanExpired removes expired entries from both caches
func (c *idAndCodeCacheImpl) CleanExpired() int {
	removedID := c.byID.CleanExpired()
	removedCode := c.byCode.CleanExpired()
	return removedID + removedCode
}

func (c *idAndCodeCacheImpl) Size() int {
	return c.byID.Size() + c.byCode.Size()
}

// GetStats returns statistics for both caches
func (c *idAndCodeCacheImpl) GetStats() []cache.CacheStats {
	idStats := c.byID.GetStats()
	idStats.Name = "IdAndCode (by ID)"

	codeStats := c.byCode.GetStats()
	codeStats.Name = "IdAndCode (by Code)"

	return []cache.CacheStats{idStats, codeStats}
}

