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
func ValidateAccess(r Resource, name string) error {
	if adminUser != name && r.Parent() != name {
		return fmt.Errorf("user '%v' does not have permission to access resource '%v'", name, r.Identity())
	}
	return nil
}
