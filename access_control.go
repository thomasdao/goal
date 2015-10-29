package goal

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

// Authorizer defines permission model for each record
type Authorizer interface {
	Authorize() Permission
}
