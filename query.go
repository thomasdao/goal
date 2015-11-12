package goal

import (
	"fmt"
	"net/http"
)

// QuerySupporter is the interface that return filtered results
// based on request paramters
type QuerySupporter interface {
	Query(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

func (api *API) queryHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(QuerySupporter); ok {
			handler = resource.Query
		}

		renderJSON(rw, request, handler)
	}
}

// AddQueryResource allows model to support query based on request
// data, return filtered results back to client
func (api *API) AddQueryResource(resource interface{}, path string) {
	api.Mux().Handle(path, api.queryHandler(resource))
}

// AddDefaultQueryPath allows model to support query based on request
// data, return filtered results back to client. The path is created
// base on struct name
func (api *API) AddDefaultQueryPath(resource interface{}) {
	queryPath := fmt.Sprintf("/query/%s/{query}", TableName(resource))
	api.AddQueryResource(resource, queryPath)
}
