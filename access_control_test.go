package goal_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/thomasdao/goal"
)

// Satisfy Roler interface
func (user *testuser) Role() []string {
	ownRole := fmt.Sprintf("testuser:%v", user.ID)
	roles := []string{ownRole}

	return roles
}

func (art *article) Permit() *goal.Permission {
	perm := &goal.Permission{}
	perm.Read = []string{"admin"}
	perm.Write = []string{"admin"}

	return perm
}

func (art *article) Get(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Read(art, request)
}

func (art *article) Post(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.Create(art, request)
}

func (art *article) Query(w http.ResponseWriter, request *http.Request) (int, interface{}, error) {
	return goal.HandleQuery(art, request)
}

func TestCanRead(t *testing.T) {
	setup()
	defer tearDown()

	// Create article with author
	author := &testuser{}
	author.Username = "secret"
	db.Create(author)

	art := &article{}
	art.author = author
	db.Create(art)

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

	artURL := fmt.Sprint(server.URL, "/article/", art.ID)

	// Make sure user is the same with current user from session
	nextReq, _ := http.NewRequest("GET", artURL, nil)
	nextReq.Header.Add("Cookie", cookies[0])

	// Get response
	client := &http.Client{}
	resp, err := client.Do(nextReq)
	fmt.Println(resp)
	resp.Body.Close()

	if err != nil {
		t.Error("Request should be unauthorized because thomasdao doesn't have admin role")
	}
}
