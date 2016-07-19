package goal_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/thomasdao/goal"
)

// Define HTTP methods to support
func (user *testuser) Get(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Read(user, request)
}

func (user *testuser) Post(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Create(user, request)
}

func (user *testuser) Put(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Update(user, request)
}

func (user *testuser) Delete(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Delete(user, request)
}

func (user *testuser) CurrentRevision() string {
	return fmt.Sprintf("%d", user.Rev)
}

func (user *testuser) SetNextRevision() {
	user.Rev = user.Rev + 1
}

func TestCreate(t *testing.T) {
	setup()
	defer tearDown()

	var json = []byte(`{"Name":"Thomas", "Age": 28, "Rev": 1}`)
	req, _ := http.NewRequest("POST", userURL(), bytes.NewBuffer(json))

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}

	if res.StatusCode != 200 {
		t.Error("Request Failed ", res.StatusCode)
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

	// Make sure data exists in Redis
	if goal.SharedCache != nil {
		key := goal.CacheKey(user)
		var redisUser testuser
		goal.SharedCache.Get(key, &redisUser)
		if !reflect.DeepEqual(user, redisUser) {
			t.Error("Incorrect data in redis, ", user, redisUser)
		}
	}
}

func TestGet(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	user.Rev = 1
	db.Create(user)

	req, _ := http.NewRequest("GET", idURL(user.ID), nil)

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

	// Make sure data exists in Redis
	if goal.SharedCache != nil {
		key := goal.CacheKey(user)

		// Test data exists in Redis
		if exist, _ := goal.SharedCache.Exists(key); !exist {
			t.Error("Data should be saved into Redis")
		}

		var redisUser testuser
		goal.SharedCache.Get(key, &redisUser)
		if !reflect.DeepEqual(user, &redisUser) {
			t.Error("Incorrect data in redis, ", user, &redisUser)
		}
	}
}

func TestPut(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	user.Rev = 1
	db.Create(user)

	var json = []byte(`{"Name":"Thomas Dao", "ID":1, "Rev":1}`)
	req, _ := http.NewRequest("PUT", idURL(user.ID), bytes.NewBuffer(json))
	req.Close = true

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		t.Errorf("Request Failed %d %s", res.StatusCode, string(body))
		return
	}

	var result testuser
	if db.Where("name = ?", "Thomas Dao").First(&result).RecordNotFound() {
		t.Error("Update unsuccessful")
	}

	if result.ID != user.ID || result.Age != user.Age || result.Rev != user.Rev+1 {
		t.Errorf("Incorrect update %+v", result)
	}

	// Make sure data exists in Redis
	if goal.SharedCache != nil {
		key := goal.CacheKey(user)
		var redisUser testuser
		goal.SharedCache.Get(key, &redisUser)
		if !reflect.DeepEqual(result, redisUser) {
			t.Error("Incorrect data in redis, ", result, redisUser)
		}
	}
}

func TestPutConflict(t *testing.T) {
	setup()
	defer tearDown()

	user := &testuser{}
	user.Name = "Thomas"
	user.Age = 28
	user.Rev = 1
	db.Create(user)

	var json = []byte(`{"Name":"Thomas Dao", "ID":1, "Rev":0}`)
	req, _ := http.NewRequest("PUT", idURL(user.ID), bytes.NewBuffer(json))
	req.Close = true

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		t.Error(err)
	}
	defer res.Body.Close()

	if res.StatusCode != 409 {
		t.Errorf("This should be conflict %d", res.StatusCode)
		return
	}

	var result testuser
	if db.Where("name = ?", "Thomas").First(&result).RecordNotFound() {
		t.Error("Update unsuccessful")
	}

	if result.ID != user.ID || result.Age != user.Age {
		t.Error("Incorrect update")
	}

	// Make sure data exists in Redis
	if goal.SharedCache != nil {
		key := goal.CacheKey(user)
		var redisUser testuser
		goal.SharedCache.Get(key, &redisUser)
		if !reflect.DeepEqual(result, redisUser) {
			t.Error("Incorrect data in redis, ", result, redisUser)
		}
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

	// Make sure no more data in redis
	if goal.SharedCache != nil {
		key := goal.CacheKey(user)
		if exist, _ := goal.SharedCache.Exists(key); exist {
			t.Error("Data should be deleted from Redis when object is deleted")
		}
	}

}
