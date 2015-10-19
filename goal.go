// Based on sleepy https://github.com/dougblack/sleepy

package goal

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/mux"
)

// An API manages a group of resources by routing requests
// to the correct method on a matching resource and marshalling
// the returned data to JSON for the HTTP response.
//
// You can instantiate multiple APIs on separate ports. Each API
// will manage its own set of resources.
type API struct {
	mux            *mux.Router
	muxInitialized bool
}

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	return &API{}
}

// To shorten the code, define a type
type simpleResponse func(http.ResponseWriter, *http.Request) (int, interface{})

// Write response back to client
func renderJson(rw http.ResponseWriter, request *http.Request, handler simpleResponse) {
	if handler == nil {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	code, data := handler(rw, request)

	content, err := json.Marshal(data)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(code)
	rw.Write(content)
}

// Mux returns Gorilla's mux.Router used by an API. If a mux
// does not yet exist, a new one will be created and returned.
func (api *API) Mux() *mux.Router {
	if api.muxInitialized {
		return api.mux
	}

	api.mux = mux.NewRouter()
	api.muxInitialized = true
	return api.mux
}

func simpleStructName(resource interface{}) string {
	// Extract name of resource type
	cls := reflect.TypeOf(resource).String()
	arr := strings.Split(cls, ".")
	name := arr[len(arr)-1]
	return name
}
