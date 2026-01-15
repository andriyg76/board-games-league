package services

import (
	"time"

	"github.com/andriyg76/bgl/cache"
	"github.com/andriyg76/bgl/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	ObjectCacheSize = 500
	ObjectCacheTTL  = 30 * time.Minute
)

// LeagueCache provides caching for League objects
type LeagueCache interface {
	Get(id primitive.ObjectID) (*models.League, bool)
	Set(id primitive.ObjectID, league *models.League)
}

type leagueCacheImpl struct {
	cache *cache.LRUCache[primitive.ObjectID, *models.League]
}

func NewLeagueCache() LeagueCache {
	return &leagueCacheImpl{
		cache: cache.NewLRUCache[primitive.ObjectID, *models.League](ObjectCacheSize, ObjectCacheTTL),
	}
}

func (c *leagueCacheImpl) Get(id primitive.ObjectID) (*models.League, bool) {
	return c.cache.Get(id)
}

func (c *leagueCacheImpl) Set(id primitive.ObjectID, league *models.League) {
	c.cache.Set(id, league, ObjectCacheTTL)
}

func (c *leagueCacheImpl) Remove(id primitive.ObjectID) {
	c.cache.Remove(id)
}

func (c *leagueCacheImpl) RemoveByKeyFunc(predicate func(primitive.ObjectID) bool) int {
	return c.cache.RemoveByKeyFunc(predicate)
}

func (c *leagueCacheImpl) RemoveByValueFunc(predicate func(*models.League) bool) int {
	return c.cache.RemoveByValueFunc(predicate)
}

func (c *leagueCacheImpl) Clear() {
	c.cache.Clear()
}

func (c *leagueCacheImpl) CleanExpired() int {
	return c.cache.CleanExpired()
}

func (c *leagueCacheImpl) Size() int {
	return c.cache.Size()
}

// GetStats returns cache statistics
func (c *leagueCacheImpl) GetStats() cache.CacheStats {
	stats := c.cache.GetStats()
	stats.Name = "League"
	return stats
}

// MembershipCacheKey is a composite key for membership cache
type MembershipCacheKey struct {
	LeagueID primitive.ObjectID
	UserID   primitive.ObjectID
}

// MembershipCache provides caching for LeagueMembership objects
type MembershipCache interface {
	Get(leagueID, userID primitive.ObjectID) (*models.LeagueMembership, bool)
	Set(leagueID, userID primitive.ObjectID, membership *models.LeagueMembership)
}

type membershipCacheImpl struct {
	cache *cache.LRUCache[MembershipCacheKey, *models.LeagueMembership]
}

func NewMembershipCache() MembershipCache {
	return &membershipCacheImpl{
		cache: cache.NewLRUCache[MembershipCacheKey, *models.LeagueMembership](ObjectCacheSize, ObjectCacheTTL),
	}
}

func (c *membershipCacheImpl) Get(leagueID, userID primitive.ObjectID) (*models.LeagueMembership, bool) {
	key := MembershipCacheKey{LeagueID: leagueID, UserID: userID}
	return c.cache.Get(key)
}

func (c *membershipCacheImpl) Set(leagueID, userID primitive.ObjectID, membership *models.LeagueMembership) {
	key := MembershipCacheKey{LeagueID: leagueID, UserID: userID}
	c.cache.Set(key, membership, ObjectCacheTTL)
}

func (c *membershipCacheImpl) Remove(leagueID, userID primitive.ObjectID) {
	key := MembershipCacheKey{LeagueID: leagueID, UserID: userID}
	c.cache.Remove(key)
}

func (c *membershipCacheImpl) RemoveByKeyFunc(predicate func(primitive.ObjectID, primitive.ObjectID) bool) int {
	return c.cache.RemoveByKeyFunc(func(key MembershipCacheKey) bool {
		return predicate(key.LeagueID, key.UserID)
	})
}

func (c *membershipCacheImpl) RemoveByValueFunc(predicate func(*models.LeagueMembership) bool) int {
	return c.cache.RemoveByValueFunc(predicate)
}

func (c *membershipCacheImpl) RemoveByLeague(leagueID primitive.ObjectID) int {
	return c.RemoveByKeyFunc(func(lID, _ primitive.ObjectID) bool {
		return lID == leagueID
	})
}

func (c *membershipCacheImpl) Clear() {
	c.cache.Clear()
}

func (c *membershipCacheImpl) CleanExpired() int {
	return c.cache.CleanExpired()
}

func (c *membershipCacheImpl) Size() int {
	return c.cache.Size()
}

func (c *membershipCacheImpl) GetStats() cache.CacheStats {
	stats := c.cache.GetStats()
	stats.Name = "Membership"
	return stats
}

// UserCache provides caching for User objects
type UserCache interface {
	GetByID(id primitive.ObjectID) (*models.User, bool)
	GetByCode(code string) (*models.User, bool)
	Set(user *models.User)
}

type userCacheImpl struct {
	byID        *cache.LRUCache[primitive.ObjectID, *models.User]
	byCode      *cache.LRUCache[string, *models.User]
	idCodeCache IdAndCodeCache
}

func NewUserCache(idCodeCache IdAndCodeCache) UserCache {
	return &userCacheImpl{
		byID:        cache.NewLRUCache[primitive.ObjectID, *models.User](ObjectCacheSize, ObjectCacheTTL),
		byCode:      cache.NewLRUCache[string, *models.User](ObjectCacheSize, ObjectCacheTTL),
		idCodeCache: idCodeCache,
	}
}

func (c *userCacheImpl) GetByID(id primitive.ObjectID) (*models.User, bool) {
	return c.byID.Get(id)
}

func (c *userCacheImpl) GetByCode(code string) (*models.User, bool) {
	if user, ok := c.byCode.Get(code); ok {
		return user, true
	}
	// Try to get by ID if we can convert code
	if idCode, err := c.idCodeCache.GetByCode(code); err == nil {
		return c.GetByID(idCode.ID)
	}
	return nil, false
}

func (c *userCacheImpl) Set(user *models.User) {
	c.byID.Set(user.ID, user, ObjectCacheTTL)
	if idCode := c.idCodeCache.GetByID(user.ID); idCode != nil {
		c.byCode.Set(idCode.Code, user, ObjectCacheTTL)
	}
}

func (c *userCacheImpl) Remove(id primitive.ObjectID) {
	c.byID.Remove(id)
	if idCode := c.idCodeCache.GetByID(id); idCode != nil {
		c.byCode.Remove(idCode.Code)
	}
}

func (c *userCacheImpl) RemoveByKeyFunc(predicate func(primitive.ObjectID) bool) int {
	return c.byID.RemoveByKeyFunc(predicate)
}

func (c *userCacheImpl) RemoveByValueFunc(predicate func(*models.User) bool) int {
	return c.byID.RemoveByValueFunc(predicate)
}

func (c *userCacheImpl) Clear() {
	c.byID.Clear()
	c.byCode.Clear()
}

func (c *userCacheImpl) CleanExpired() int {
	removedID := c.byID.CleanExpired()
	removedCode := c.byCode.CleanExpired()
	return removedID + removedCode
}

func (c *userCacheImpl) Size() int {
	return c.byID.Size() + c.byCode.Size()
}

func (c *userCacheImpl) GetStats() []cache.CacheStats {
	idStats := c.byID.GetStats()
	idStats.Name = "User (by ID)"

	codeStats := c.byCode.GetStats()
	codeStats.Name = "User (by Code)"

	return []cache.CacheStats{idStats, codeStats}
}
