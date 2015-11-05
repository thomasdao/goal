package goal

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"

	"github.com/gorilla/sessions"
)

const (
	// SessionName is default name for user session
	SessionName = "goal.UserSessionName"

	// SessionKey is default key for user object
	SessionKey = "goal.UserSessionKey"
)

// SharedSessionStore used to generate session for multiple requests
var SharedSessionStore sessions.Store

// InitSessionStore initializes SharedSessionStore
func InitSessionStore(store sessions.Store) {
	SharedSessionStore = store
}

var userType reflect.Type

// RegisterUserModel lets goal which model act as user
func RegisterUserModel(user interface{}) {
	userType = reflect.TypeOf(user).Elem()
}

// getUserResource returns a new variable based on reflection
// e.g var user User
func getUserResource() (interface{}, error) {
	if userType == nil {
		return nil, errors.New("User model was not registered")
	}

	return reflect.Zero(userType).Interface(), nil
}

// SetUserSession sets current user to session
func SetUserSession(w http.ResponseWriter, req *http.Request) error {
	session, err := SharedSessionStore.Get(req, SessionName)
	if err != nil {
		return err
	}

	var user interface{}
	user, err = getUserResource() // similar to var user User
	if err != nil {
		return err
	}

	// Marshal user to json and save to session
	var data []byte
	data, err = json.Marshal(&user)
	if err != nil {
		return err
	}

	// Set some session values.
	session.Values[SessionKey] = data

	// Save it before we write to the response/return from the handler.
	err = session.Save(req, w)
	return err
}

// GetCurrentUser returns current user based on the request header
func GetCurrentUser(req *http.Request) (interface{}, error) {
	session, err := SharedSessionStore.Get(req, SessionName)
	if err != nil {
		return nil, err
	}

	data := session.Values[SessionKey]

	var user interface{}
	user, err = getUserResource()
	if err != nil {
		return nil, err
	}

	if data, ok := data.([]byte); ok {
		err = json.Unmarshal(data, &user)
		return &user, err
	}

	return nil, errors.New("invalid session data")
}

// ClearUserSession removes the current user from session
func ClearUserSession(w http.ResponseWriter, req *http.Request) error {
	session, err := SharedSessionStore.Get(req, SessionName)
	if err != nil {
		return err
	}

	delete(session.Values, SessionKey)
	err = session.Save(req, w)
	return err
}
