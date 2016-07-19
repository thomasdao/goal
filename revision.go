package goal

// Revisioner tell the revision of current record. It is
// typically used to avoid data being overridden by multiple clients
type Revisioner interface {
	CurrentRevision() string
	SetNextRevision()
}

// CanMerge check if the updated object can be safely merged to current
// object
func CanMerge(current Revisioner, updated Revisioner) bool {
	// If the revision of updated object is the same as current revision,
	// updated object can be merged safely
	if current.CurrentRevision() == updated.CurrentRevision() {
		return true
	}

	return false
}
