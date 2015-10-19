package goal

import (
	"fmt"
	"net/http"
)

// HTTP Methods
const (
	GET    = "GET"
	POST   = "POST"
	PUT    = "PUT"
	DELETE = "DELETE"
	HEAD   = "HEAD"
	PATCH  = "PATCH"
)

// GetSupporter is the interface that provides the Get
// method a resource must support to receive HTTP GETs.
type GetSupporter interface {
	Get(http.ResponseWriter, *http.Request) (int, interface{})
}

// PostSupporter is the interface that provides the Post
// method a resource must support to receive HTTP POSTs.
type PostSupporter interface {
	Post(http.ResponseWriter, *http.Request) (int, interface{})
}

// PutSupporter is the interface that provides the Put
// method a resource must support to receive HTTP PUTs.
type PutSupporter interface {
	Put(http.ResponseWriter, *http.Request) (int, interface{})
}

// DeleteSupporter is the interface that provides the Delete
// method a resource must support to receive HTTP DELETEs.
type DeleteSupporter interface {
	Delete(http.ResponseWriter, *http.Request) (int, interface{})
}

// HeadSupporter is the interface that provides the Head
// method a resource must support to receive HTTP HEADs.
type HeadSupporter interface {
	Head(http.ResponseWriter, *http.Request) (int, interface{})
}

// PatchSupporter is the interface that provides the Patch
// method a resource must support to receive HTTP PATCHs.
type PatchSupporter interface {
	Patch(http.ResponseWriter, *http.Request) (int, interface{})
}

// Route request to correct handler and write result back to client
func (api *API) crudHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		switch request.Method {
		case GET:
			if resource, ok := resource.(GetSupporter); ok {
				handler = resource.Get
			}
		case POST:
			if resource, ok := resource.(PostSupporter); ok {
				handler = resource.Post
			}
		case PUT:
			if resource, ok := resource.(PutSupporter); ok {
				handler = resource.Put
			}
		case DELETE:
			if resource, ok := resource.(DeleteSupporter); ok {
				handler = resource.Delete
			}
		case HEAD:
			if resource, ok := resource.(HeadSupporter); ok {
				handler = resource.Head
			}
		case PATCH:
			if resource, ok := resource.(PatchSupporter); ok {
				handler = resource.Patch
			}
		}

		renderJSON(rw, request, handler)
	}
}

// AddCrudResource adds a new resource to an API. The API will route
// requests that match one of the given paths to the matching HTTP
// method on the resource.
func (api *API) AddCrudResource(resource interface{}, paths ...string) {
	for _, path := range paths {
		api.Mux().HandleFunc(path, api.crudHandler(resource))
	}
}

// AddDefaultCrudPaths adds default path for a resource.
// The default path is based on the struct name
func (api *API) AddDefaultCrudPaths(resource interface{}) {
	// Extract name of resource type
	name := simpleStructName(resource)

	// Default path to interact with resource
	createPath := fmt.Sprintf("/%s", name)
	detailPath := fmt.Sprintf("/%s/{id:[0-9]+}", name)

	api.AddCrudResource(resource, createPath, detailPath)
}
