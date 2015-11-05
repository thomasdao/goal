package goal

import "net/http"

// Registerer register a new user to system
type Registerer interface {
	Register(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// Loginer authenticates user into the system
type Loginer interface {
	Login(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

// Logouter clear sessions and log user out
type Logouter interface {
	Logout(http.ResponseWriter, *http.Request) (int, interface{}, error)
}

func (api *API) registerHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Registerer); ok {
			handler = resource.Register
		}

		renderJSON(rw, request, handler)
	}
}

func (api *API) loginHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Loginer); ok {
			handler = resource.Login
		}

		renderJSON(rw, request, handler)
	}
}

func (api *API) logoutHandler(resource interface{}) http.HandlerFunc {
	return func(rw http.ResponseWriter, request *http.Request) {
		var handler simpleResponse

		if resource, ok := resource.(Logouter); ok {
			handler = resource.Logout
		}

		renderJSON(rw, request, handler)
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

// AddLogoutPath let user logout from the system
func (api *API) AddLogoutPath(resource interface{}, path string) {
	api.Mux().Handle(path, api.logoutHandler(resource))
}

// AddDefaultAuthPaths route request to the model which implement
// authentications
func (api *API) AddDefaultAuthPaths(resource interface{}) {
	api.Mux().Handle("/auth/register", api.registerHandler(resource))
	api.Mux().Handle("/auth/login", api.loginHandler(resource))
	api.Mux().Handle("/auth/logout", api.logoutHandler(resource))
}
