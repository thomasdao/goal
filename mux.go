// Based on sleepy https://github.com/dougblack/sleepy

package goal

import "github.com/gorilla/mux"

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

var sharedAPI *API

// NewAPI allocates and returns a new API.
func NewAPI() *API {
	if sharedAPI == nil {
		sharedAPI = &API{}
	}
	return sharedAPI
}

// SharedAPI return API instance
func SharedAPI() *API {
	return sharedAPI
}

// Mux returns Gorilla's mux.Router used by an API. If a mux
// does not yet exist, a new one will be created and returned
func (api *API) Mux() *mux.Router {
	if api.muxInitialized {
		return api.mux
	}

	api.mux = mux.NewRouter()
	api.muxInitialized = true
	return api.mux
}
