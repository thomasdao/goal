package goal_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
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

	res := httptest.NewRecorder()

	var json = []byte(`{"username":"thomasdao", "password": "something-secret"}`)
	req, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(json))
	req.Header.Set("Content-Type", "application/json")

	goal.SharedAPI().Mux().ServeHTTP(res, req)

	// Make sure cookies is set properly
	hdr := res.Header()
	cookies, ok := hdr["Set-Cookie"]
	if !ok || len(cookies) != 1 {
		t.Fatal("No cookies. Header:", hdr)
	}

	// Make sure db has one object
	var user testuser
	db.Where("username = ?", "thomasdao").First(&user)
	if &user == nil {
		t.Error("Fail to save object to database")
		return
	}

	// Make sure user is the same with current user from session
	nextReq, _ := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(json))
	nextReq.Header.Add("Cookie", cookies[0])
	currentUser, err := goal.GetCurrentUser(nextReq)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(&user, currentUser) {
		t.Error("Get invalid current user from request")
	}
}
