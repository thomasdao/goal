package goal_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/thomasdao/goal"
)

func (user *testuser) Query(w http.ResponseWriter, req *http.Request) (int, interface{}) {
	var results []testuser
	return goal.HandleQuery(user, req, &results)
}

func queryPath(query []byte) string {
	return fmt.Sprint(server.URL, "/query/testuser/", url.QueryEscape(string(query)))
}

func createUsers() {
	// Prepare data
	names := []string{"Thomas", "Alan", "Jason", "Ben"}
	ages := []int{28, 30, 22, 40}
	for index, name := range names {
		user := &testuser{}
		user.Name = name
		user.Age = ages[index]
		db.Create(user)
	}
}

func TestSuccessQueryParamsFind(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()

	// Test success case

	item := goal.QueryItem{}
	item.Key = "name"
	item.Op = "="
	item.Val = "Thomas"

	params := &goal.QueryParams{}
	params.Where = []goal.QueryItem{item}

	var results []testuser
	var user testuser
	params.Find(&user, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result")
	}

	orItem := goal.QueryItem{}
	orItem.Key = "name"
	orItem.Op = "="
	orItem.Val = "Alan"
	item.Or = []goal.QueryItem{orItem}
	params.Where = []goal.QueryItem{item}
	params.Find(&user, &results)

	if results == nil || len(results) != 2 {
		t.Error("Error: query should return 2 result")
	}

	andItem := goal.QueryItem{}
	andItem.Key = "age"
	andItem.Op = ">"
	andItem.Val = "29"
	params.Where = []goal.QueryItem{item, andItem}
	params.Find(&user, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result")
	}

}

func TestInvalidQueryParamsFind(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()

	// Test success case

	item := goal.QueryItem{}
	item.Key = "name"
	item.Op = "hello"
	item.Val = "Thomas"

	params := &goal.QueryParams{}
	params.Where = []goal.QueryItem{item}

	var results []testuser
	var user testuser
	err := params.Find(&user, &results)
	if err == nil {
		t.Error("Error: Query operator should be invalid")
	}

	item = goal.QueryItem{}
	item.Key = "hello"
	item.Op = "="
	item.Val = "Thomas"
	params.Where = []goal.QueryItem{item}

	err = params.Find(&user, &results)
	if err == nil {
		t.Error("Error: Query column should be invalid")
	}
}

func TestQueryViaAPI(t *testing.T) {
	setup()
	defer tearDown()

	createUsers()
	item := goal.QueryItem{}
	item.Key = "name"
	item.Op = "="
	item.Val = "Thomas"

	params := &goal.QueryParams{}
	params.Where = []goal.QueryItem{item}

	query, _ := json.Marshal(params)

	req, _ := http.NewRequest("GET", queryPath(query), nil)
	req.Header.Set("Content-Type", "application/json")

	// Get response
	client := &http.Client{}
	res, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		t.Error(err)
	}

	if res.StatusCode != 200 {
		fmt.Println(res.StatusCode)
		t.Error("Request Failed")
		return
	}

	defer res.Body.Close()
	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Error(err)
		return
	}

	var results []testuser
	json.Unmarshal(content, &results)

	if results == nil || len(results) != 1 {
		t.Error("Error: query should return 1 result")
	}
}
