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

// PermitReader allows authenticated user to read the record
type PermitReader interface {
	PermitRead() []string
}

// PermitWriter allows authenticated user to write the record
type PermitWriter interface {
	PermitWrite() []string
}

// CanPerform check if a roler can access a resource (read/write)
// If read is false, then it will check for write permission
// It will return error if the check is failed
func CanPerform(resource interface{}, request *http.Request, read bool) error {
	unauthorized := errors.New("unauthorized access")

	// If a resource does not define PermitRead and PermitWrite method,
	// we assume it is public.
	var roles []string
	if read {
		permitReader, ok := resource.(PermitReader)
		if ok {
			roles = permitReader.PermitRead()
		}
	} else {
		permitWriter, ok := resource.(PermitWriter)
		if ok {
			roles = permitWriter.PermitWrite()
		}
	}

	if len(roles) == 0 {
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
