package auth

import (
	"fmt"

	"github.com/google/uuid"
)

const adminUser = "admin"

type Resource interface {
	Parent() string
	Identity() uuid.UUID
}

// Validate the given names have access to the given resource by checking
// if one of the names is [adminUser] or if the parent matches one of the given names.
func ValidateAccess(r Resource, names []string) error {
	parent := r.Parent()

	for _, name := range names {
		if parent == name || adminUser == name {
			return nil
		}
	}

	return fmt.Errorf("user '%v' does not have permission to access resource '%v'", names, r.Identity())
}
