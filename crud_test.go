package goal_test

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garyburd/redigo/redis"
	"github.com/jinzhu/gorm"
	"github.com/thomasdao/goal"
)

var server *httptest.Server

type testuser struct {
	ID   uint `gorm:"primary_key"`
	Name string
	Age  int
}

func (user *testuser) Get(w http.ResponseWriter, request *http.Request) (int, interface{}) {
	return goal.Read(user, request)
}

func (user *testuser) Post(w http.ResponseWriter, request *http.Request) (int, interface{}) {
	return goal.Create(user, request)
}

func (user *testuser) Put(w http.ResponseWriter, request *http.Request) (int, interface{}) {
	return goal.Update(user, request)
}

func (user *testuser) Delete(w http.ResponseWriter, request *http.Request) (int, interface{}) {
	return goal.Delete(user, request)
}

func (user *testuser) AfterSave() error {
	goal.RedisSet(user)
	return nil
}

func (user *testuser) BeforeDelete() error {
	fmt.Printf("Before Delete %s", user.ID)
	goal.RedisUnset(user)
	return nil
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

	goal.InitRedisPool(pool)

	// Initialize API
	api := goal.NewAPI()

	// Initialize resource
	var user testuser
	db.AutoMigrate(user)

	// Add default path
	api.AddDefaultCrudPaths(&user)

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
		goal.Pool().Close()
	}
}

func userURL() string {
	return fmt.Sprint(server.URL, "/testuser")
}

func idURL(id uint) string {
	return fmt.Sprint(server.URL, "/testuser/", id)
}

func TestCreate(t *testing.T) {
	setup()
	defer tearDown()

	var json = []byte(`{"Name":"Thomas", "Age": 28}`)
	req, _ := http.NewRequest("POST", userURL(), bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Request Failed")
		return
	}

	// Make sure db has one object
	var user testuser
	db.Where("name = ?", "Thomas").First(&user)
	if &user == nil {
		t.Error("Fail to save object to database")
		return
	}

	if user.Name != "Thomas" || user.Age != 28 {
		t.Error("Save wrong data or missing data")
	}
}

func TestGet(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	db.Create(user)

	req, _ := http.NewRequest("GET", idURL(user.ID), nil)
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Request Failed")
		return
	}

	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var result testuser
	err = json.Unmarshal(content, &result)

	if err != nil {
		t.Error(err)
		return
	}

	if result.ID != user.ID || result.Name != user.Name || result.Age != user.Age {
		t.Error("Response is invalid")
	}
}

func TestPut(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	db.Create(user)

	var json = []byte(`{"Name":"Thomas Dao"}`)
	req, _ := http.NewRequest("PUT", idURL(user.ID), bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Request Failed")
		return
	}

	var result testuser
	if db.Where("name = ?", "Thomas Dao").First(&result).RecordNotFound() {
		t.Error("Update unsuccessful")
	}

	if result.ID != user.ID || result.Age != user.Age {
		t.Error("Incorrect update")
	}
}

func TestDelete(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	db.Create(user)

	req, _ := http.NewRequest("DELETE", idURL(user.ID), nil)
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Request Failed")
		return
	}

	var result testuser
	if !db.Where("name = ?", "Thomas").First(&result).RecordNotFound() {
		t.Error("Delete is not successful. Expected result delete from db")
	}

}
