package goal

import (
	"fmt"
	"net/http"
)

// *********************************************************************//
// For Query
// *********************************************************************//

// QuerySupporter is the interface that return filtered results
// based on request paramters
type QuerySupporter interface {
	Query(*http.Request) (int, interface{})
}

func (api *API) queryHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(QuerySupporter); ok {
			handler = resource.Query
		}

		writeResponse(rw, request, handler)
	}
}

// AddQueryPath allows model to support query based on request
// data, return filtered results back to client
func (api *API) AddQueryPath(resource interface{}, path string) {
	api.Mux().Handle(path, api.queryHandler(resource))
}

// AddDefaultQueryPath allows model to support query based on request
// data, return filtered results back to client. The path is created
// base on struct name
func (api *API) AddDefaultQueryPath(resource interface{}) {
	queryPath := fmt.Sprintf("/query/%s", simpleStructName(resource))
	api.AddQueryPath(resource, queryPath)
}
