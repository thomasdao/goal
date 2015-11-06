package goal_test

import (
	"bytes"
	"net/http"
	"reflect"
	"testing"

	"github.com/thomasdao/goal"
)

// Setup methods to conform to auth interfaces
func (user *testuser) Register(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := goal.RegisterWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Login(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	currentUser, err := goal.LoginWithPassword(w, req, "username", "password")

	if err != nil {
		return 500, nil, err
	}

	return 200, currentUser, nil
}

func (user *testuser) Logout(w http.ResponseWriter, req *http.Request) (int, interface{}, error) {
	goal.HandleLogout(w, req)
	return 200, nil, nil
}

func TestRegister(t *testing.T) {
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
	if goal.Pool() != nil {
		key := goal.RedisKey(user)
		var redisUser testuser
		goal.RedisGet(key, &redisUser)
		if !reflect.DeepEqual(user, redisUser) {
			t.Error("Incorrect data in redis, ", user, redisUser)
		}
	}
}
