package goal

import "net/http"

// Registerer register a new user to system
type Registerer interface {
	Register(*http.Request) (int, interface{})
}

// Loginer authenticates user into the system
type Loginer interface {
	Login(*http.Request) (int, interface{})
}

func (api *API) registerHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Registerer); ok {
			handler = resource.Register
		}

		writeResponse(rw, request, handler)
	}
}

func (api *API) loginHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Loginer); ok {
			handler = resource.Login
		}

		writeResponse(rw, request, handler)
	}
}

// AddRegisterPath let user to register into a system
func (api *API) AddRegisterPath(resource interface{}, path string) {
	api.Mux().Handle(path, api.registerHandler(resource))
}

// AddLoginPath let user login to system
func (api *API) AddLoginPath(resource interface{}, path string) {
	api.Mux().Handle(path, api.loginHandler(resource))
}

// AddDefaultAuthPaths route request to the model which implement
// authentications
func (api *API) AddDefaultAuthPaths(resource interface{}) {
	api.Mux().Handle("/auth/register", api.registerHandler(resource))
	api.Mux().Handle("/auth/login", api.loginHandler(resource))
}
