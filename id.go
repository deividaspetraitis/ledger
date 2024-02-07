package ledger

import "github.com/google/uuid"

// newID generates new unique ID.
func newID() string {
	return uuid.NewString()
}

// isValidID reports whether given ID string is valid ID.
func isValidID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
