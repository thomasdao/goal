package goal

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

// Cacher defines a interface for fast key-value caching
type Cacher interface {
	Get(string, interface{}) error
	Set(string, interface{}) error
	Delete(string) error
	Exists(string) (bool, error)
}

// SharedCache is global variable to cache data
var SharedCache Cacher

// RegisterCacher initializes SharedCache
func RegisterCacher(cache Cacher) {
	SharedCache = cache

	if SharedCache != nil {
		// Register Gorm callbacks
		if db != nil {
			db.Callback().Create().After("gorm:after_create").Register("goal:cache_after_create", Cache)
			db.Callback().Update().After("gorm:after_update").Register("goal:cache_after_update", Cache)
			db.Callback().Query().After("gorm:after_query").Register("goal:cache_after_query", Cache)
			db.Callback().Delete().Before("gorm:before_delete").Register("goal:uncache_after_delete", Uncache)
		}
	}
}

func cacheKeyFromScope(scope *gorm.Scope) string {
	name := scope.TableName()
	id := scope.PrimaryKeyValue()
	key := DefaultCacheKey(name, id)
	return key
}

// CacheKey defines by the struct or fallback
// to name:id format
func CacheKey(resource interface{}) string {
	scope := db.NewScope(resource)
	return cacheKeyFromScope(scope)
}

// DefaultCacheKey returns default format for redis key
func DefaultCacheKey(name string, id interface{}) string {
	return fmt.Sprintf("%v:%v", name, id)
}

// Uncache data from cacher
func Uncache(scope *gorm.Scope) {
	// Reload object before delete
	scope.DB().New().First(scope.Value)

	// Delete from redis
	key := cacheKeyFromScope(scope)
	SharedCache.Delete(key)
}

// Cache data to cacher
func Cache(scope *gorm.Scope) {
	key := cacheKeyFromScope(scope)
	SharedCache.Set(key, scope.Value)
}
