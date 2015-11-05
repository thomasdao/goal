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

	// Retrieve role from current user
	user, err := GetCurrentUser(request)
	if err != nil {
		return err
	}

	var roler Roler
	if user != nil {
		roler, _ = user.(Roler)
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
