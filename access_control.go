package goal

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Roler is usually assigned to User class, which define which
// role user has: ["admin", "user_id"]
type Roler interface {
	Roles() []string
}

// PermitReader allows authenticated user to read the record
type PermitReader interface {
	PermitRead() []string
}

// PermitWriter allows authenticated user to write the record
type PermitWriter interface {
	PermitWrite() []string
}

// Permission makes it easier to implement access control
type Permission struct {
	Read  string
	Write string
}

// PermitRead conforms to PermitReader interface
func (p *Permission) PermitRead() []string {
	if p.Read != "" {
		var roles []string
		err := json.Unmarshal([]byte(p.Read), &roles)
		if err != nil {
			return nil
		}

		return roles
	}
	return nil
}

// PermitWrite conforms to PermitWriter interface
func (p *Permission) PermitWrite() []string {
	if p.Write != "" {
		var roles []string
		err := json.Unmarshal([]byte(p.Write), &roles)
		if err != nil {
			return nil
		}

		return roles
	}

	return nil
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
		for _, role := range roler.Roles() {
			if change == role {
				return nil
			}
		}
	}

	return unauthorized
}
