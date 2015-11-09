// Cache object to redis automatically by registering
// callback to gorm

package goal

import (
	"encoding/json"
	"fmt"

	"github.com/garyburd/redigo/redis"
)

// RedisCache implements Cacher interface
type RedisCache struct{}

// Get returns data for a key
func (cache *RedisCache) Get(key string, val interface{}) error {
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
	json.Unmarshal(reply, val)

	return nil
}

// Set a val for a key into Redis
func (cache *RedisCache) Set(key string, val interface{}) error {
	conn, err := pool.Dial()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer conn.Close()

	var data []byte
	data, err = json.Marshal(val)
	if err != nil {
		return err
	}

	_, err = conn.Do("SET", key, data)
	return err
}

// Delete a key from Redis
func (cache *RedisCache) Delete(key string) error {
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

// Exists checks if a key exists inside Redis
func (cache *RedisCache) Exists(key string) (bool, error) {
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

var pool *redis.Pool

// InitRedisPool initializes Redis and connection pool
func (cache *RedisCache) InitRedisPool(p *redis.Pool) error {
	pool = p

	conn, err := pool.Dial()
	if err != nil {
		pool = nil
		return err
	}

	defer conn.Close()
	return nil
}

// Pool returns global connection pool
func Pool() *redis.Pool {
	return pool
}

// RedisClearAll clear all data from connection's CURRENT database
func RedisClearAll() error {
	if pool == nil {
		return nil
	}
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
