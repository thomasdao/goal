package goal

import (
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

// SetUserModel lets goal which model act as user
func SetUserModel(user interface{}) {
	userType = reflect.TypeOf(user).Elem()
}

// getUserResource returns a new variable based on reflection
// e.g user := &User{}
func getUserResource() (interface{}, error) {
	if userType == nil {
		return nil, errors.New("User model was not registered")
	}

	return reflect.New(userType).Interface(), nil
}

// SetUserSession sets current user to session
func SetUserSession(w http.ResponseWriter, req *http.Request, user interface{}) error {
	session, err := SharedSessionStore.Get(req, SessionName)
	if err != nil {
		return err
	}

	scope := db.NewScope(user)

	// Set some session values.
	session.Values[SessionKey] = scope.PrimaryKeyValue()

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

	userID, ok := session.Values[SessionKey]
	if !ok {
		return nil, errors.New("empty session")
	}

	var user interface{}
	user, err = getUserResource()
	if err != nil {
		return nil, err
	}

	// Load user from Cache or from database
	exists := false
	if SharedCache != nil {
		cacheKey := DefaultCacheKey(TableName(user), userID)
		exists, err = SharedCache.Exists(cacheKey)
		if err == nil && exists {
			err = SharedCache.Get(cacheKey, user)

			if err == nil {
				return user, nil
			}
		}
	}

	// If data not exists in Redis, load from database
	if !exists {
		err = db.First(user, userID).Error
		return user, err
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
