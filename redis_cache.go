// Cache object to redis automatically by registering
// callback to gorm

package goal

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
)

var pool *redis.Pool

// InitRedisPool initialize connection pool to redis
func InitRedisPool(p *redis.Pool) {
	pool = p

	conn, err := pool.Dial()
	if err != nil {
		pool = nil
		return
	}

	defer conn.Close()

	// Register Gorm callbacks
	if db != nil {
		db.Callback().Create().After("gorm:after_create").Register("goal:cache_after_create", cacheToRedis)
		db.Callback().Update().After("gorm:after_update").Register("goal:cache_after_update", cacheToRedis)
		db.Callback().Query().After("gorm:after_query").Register("goal:cache_after_query", cacheToRedis)
		db.Callback().Delete().Before("gorm:before_delete").Register("goal:uncache_after_delete", uncacheFromRedis)
	}
}

// Pool return global connection pool to redis
func Pool() *redis.Pool {
	return pool
}

func uncacheFromRedis(scope *gorm.Scope) {
	// Reload object before delete
	scope.DB().New().First(scope.Value)

	// Delete from redis
	key := redisKeyFromScope(scope)
	RedisUnset(key)
}

func cacheToRedis(scope *gorm.Scope) {
	key := redisKeyFromScope(scope)
	RedisSet(key, scope.Value)
}

func redisKeyFromScope(scope *gorm.Scope) string {
	name := scope.TableName()
	id := scope.PrimaryKeyValue()
	key := DefaultRedisKey(name, id)
	return key
}

// RedisKey defines by the struct or fallback
// to name:id format
func RedisKey(resource interface{}) string {
	scope := db.NewScope(resource)
	return redisKeyFromScope(scope)
}

// DefaultRedisKey returns default format for redis key
func DefaultRedisKey(name string, id interface{}) string {
	return fmt.Sprintf("%v:%v", name, id)
}

// RedisSet saves object to redis
func RedisSet(key string, resource interface{}) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	var val []byte
	val, err = json.Marshal(resource)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, val)
	return err
}

// RedisUnset delete object from redis
func RedisUnset(key string) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	fmt.Printf("UNSET %s", key)
	_, err = conn.Do("DEL", key)
	return err
}

// RedisExists check if a key exists
func RedisExists(key string) (bool, error) {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	defer conn.Close()

	var reply bool
	reply, err = redis.Bool(conn.Do("EXISTS", key))

	return reply, err
}

// RedisGet returns object from database
func RedisGet(key string, resource interface{}) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	if key == "" {
		return nil
	}

	var reply []byte
	reply, err = redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}

	// Populate resource
	json.Unmarshal(reply, resource)

	return nil
}

// RedisClearAll clear all data from connection's CURRENT database
func RedisClearAll() error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	_, err = conn.Do("FLUSHDB")

	if err != nil {
		fmt.Println("Error clear redis ", err)
	}

	return err
}
