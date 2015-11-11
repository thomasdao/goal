package goal_test

import (
	"flag"
	"fmt"
	"net/http/httptest"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/sessions"
	"github.com/jinzhu/gorm"
	"github.com/thomasdao/goal"
)

var server *httptest.Server

type testuser struct {
	ID       uint `gorm:"primary_key"`
	Username string
	Password string
	Name     string
	Age      int
}

type article struct {
	ID     uint `gorm:"primary_key"`
	author *testuser
}

var db gorm.DB

var (
	redisAddress   = flag.String("redis-address", ":6379", "Address to the Redis server")
	maxConnections = flag.Int("max-connections", 10, "Max connections to Redis")
)

func setup() {
	var err error
	db, err = gorm.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}

	db.SingularTable(true)

	// Setup database
	goal.InitGormDb(&db)

	// Setup redis
	pool := redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *redisAddress)

		if err != nil {
			return nil, err
		}

		return c, err
	}, *maxConnections)

	redisCache := &goal.RedisCache{}
	err = redisCache.InitRedisPool(pool)
	if err == nil {
		goal.RegisterCacher(redisCache)
	}

	// Initialize API
	api := goal.NewAPI()

	// Initialize resource
	var user testuser

	models := []interface{}{&user, &article{}}

	// Add default path
	for _, model := range models {
		goal.RegisterModel(model)
	}

	goal.SetUserModel(&user)

	store := sessions.NewCookieStore([]byte("something-very-secret"))
	goal.InitSessionStore(store)

	// Setup testing server
	server = httptest.NewServer(api.Mux())
}

func tearDown() {
	if server != nil {
		server.Close()
	}

	if goal.DB() != nil {
		db.Close()
	}

	if goal.Pool() != nil {
		goal.RedisClearAll()
		goal.Pool().Close()
	}
}

func userURL() string {
	return fmt.Sprint(server.URL, "/testuser")
}

func idURL(id interface{}) string {
	return fmt.Sprint(server.URL, "/testuser/", id)
}
