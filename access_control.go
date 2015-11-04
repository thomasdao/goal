package goal

import (
	"errors"
	"net/http"
)

// Roler is usually assigned to User class, which define which
// role user has: ["admin", "user_id"]
type Roler interface {
	Role() []string
}

// Permission defines which role has read or write access
type Permission struct {
	Read  []string
	Write []string
}

// Permitter defines permission model for each record
type Permitter interface {
	Permit() *Permission
}

// RequestRoler defines an object to return current role of a request
type RequestRoler interface {
	CurrentRole(*http.Request) Roler
}

// SharedRequestRoler defines single instance of role manager
var SharedRequestRoler RequestRoler

// SetSharedRequestRoler initializes the share manager
func SetSharedRequestRoler(manager RequestRoler) {
	SharedRequestRoler = manager
}

// CanPerform check if a roler can access a resource (read/write)
// If read is false, then it will check for write permission
// It will return error if the check is failed
func CanPerform(resource interface{}, request *http.Request, read bool) error {
	permitter, ok := resource.(Permitter)

	// If resource does not define permission model, it assumes can
	// be interacted by public
	if !ok {
		return nil
	}

	permission := permitter.Permit()

	if permission == nil {
		return nil
	}

	var roler Roler
	if SharedRequestRoler != nil {
		roler = SharedRequestRoler.CurrentRole(request)
	} else {
		roler = nil
	}

	var roles []string
	if read {
		roles = permission.Read
	} else {
		roles = permission.Write
	}

	unauthorized := errors.New("unauthorized access")

	// If roles is not defined, then this resource does not allow that action
	if roler == nil || roles == nil {
		return unauthorized
	}

	// Check if roler has role inside permision
	for _, change := range roles {
		for _, role := range roler.Role() {
			if change == role {
				return nil
			}
		}
	}

	return unauthorized
}
