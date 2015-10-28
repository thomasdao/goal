// Define data structure for a query request
// {
//   "where":[{"key": "name", "op": "=", "val": "Thomas"}],
//   "order": [{"key": "name", "val": "asc"}]
//   "limit": 1
// }

package goal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

var ops map[string]bool

func allowedOps() map[string]bool {
	if ops == nil {
		ops = map[string]bool{
			"=":    true,
			">":    true,
			">=":   true,
			"<":    true,
			"<=":   true,
			"<>":   true,
			"in":   true,
			"like": true,
		}
	}

	return ops
}

// QueryItem defines most basic element of a query.
// For example: name = Thomas
type QueryItem struct {
	Key string      `json:"key"`
	Op  string      `json:"op"`
	Val interface{} `json:"val"`
	Or  []QueryItem `json:"or"`
}

func (item *QueryItem) getQuery(scope *gorm.Scope) (string, error) {
	_, exists := allowedOps()[item.Op]
	if !exists {
		return "", errors.New("Invalid SQL operator")
	}

	if !scope.HasColumn(item.Key) {
		str := fmt.Sprintf("Column %s does not exist", item.Key)
		return "", errors.New(str)
	}

	var query string

	if item.Op == "in" {
		query = fmt.Sprintf("%s %s (?)", item.Key, item.Op)
	} else {
		query = fmt.Sprintf("%s %s ?", item.Key, item.Op)
	}

	return query, nil
}

// QueryParams defines structure of a query. Where clause
// may include multiple QueryItem and connect by "AND" operator
type QueryParams struct {
	Where []QueryItem     `json:"where"`
	Limit int64           `json:"limit"`
	Order map[string]bool `json:"order"`
}

// Find constructs the query, return error immediately if query is invalid,
// and query database if everything is valid
func (params *QueryParams) Find(resource interface{}, results interface{}) error {
	scope := db.NewScope(resource)

	qryDB := db.New()

	// Parse where clause
	if params.Where != nil {
		for _, item := range params.Where {
			query, err := item.getQuery(scope)

			// Return immediately if query is invalid
			if err != nil {
				return err
			}

			qryDB = qryDB.Where(query, item.Val)

			if item.Or != nil {
				for _, orItem := range item.Or {
					query, err = orItem.getQuery(scope)

					// Return immediately if query is invalid
					if err != nil {
						return err
					}

					qryDB = qryDB.Or(query, orItem.Val)
				}
			}
		}
	}

	if params.Limit != 0 {
		qryDB = qryDB.Limit(params.Limit)
	}

	if params.Order != nil {
		for name, order := range params.Order {
			if !scope.HasColumn(name) {
				errorMsg := fmt.Sprintf("Column %s does not exist", name)
				return errors.New(errorMsg)
			}

			qryDB = qryDB.Order(name, order)
		}
	}

	// Query the database
	qryDB.Find(&results)

	return nil
}

// Query retrieves parameters from request and construct a proper SQL query
func Query(resource interface{}, request *http.Request) (int, interface{}) {
	if db == nil {
		panic("Database is not initialized yet")
	}

	vars := mux.Vars(request)

	// Retrieve query parameter
	query := vars["query"]

	var params QueryParams
	json.Unmarshal([]byte(query), &params)

	var results interface{}
	params.Find(resource, results)

	return 200, results
}
