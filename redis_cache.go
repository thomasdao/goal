// Cache object to redis. It should be called from
// AfterSave and AfterDelete callbacks

package goal

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/garyburd/redigo/redis"
)

// Cacher let implementer define it own cache key
type Cacher interface {
	CacheKey() string
}

var pool *redis.Pool

// InitRedisPool initialize connection pool to redis
func InitRedisPool(p *redis.Pool) {
	pool = p
}

// Pool return global connection pool to redis
func Pool() *redis.Pool {
	return pool
}

// Get redis key as defined by the struct or fallback
// to name:id format
func redisKey(resource interface{}) string {
	if resource == nil {
		return ""
	}

	var key string

	if cacher, ok := resource.(Cacher); ok {
		key = cacher.CacheKey()
	} else {
		name := simpleStructName(resource)
		val := reflect.ValueOf(resource).Elem()
		id := val.FieldByName("ID").Interface()
		key = DefaultRedisKey(name, id)
	}

	fmt.Printf("Redis Key :%s", key)

	return key
}

func RedisKey(resource interface{}) string {
	return redisKey(resource)
}

// DefaultRedisKey returns default format for redis key
func DefaultRedisKey(name string, id interface{}) string {
	return fmt.Sprintf("%s:%d", name, id)
}

// RedisSet saves object to redis
func RedisSet(resource interface{}) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	key := redisKey(resource)
	var val []byte
	val, err = json.Marshal(resource)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, val)
	return err
}

// RedisUnset delete object from redis
func RedisUnset(resource interface{}) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	key := redisKey(resource)
	fmt.Printf("UNSET %s", key)
	_, err = conn.Do("DEL", key)
	return err
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
