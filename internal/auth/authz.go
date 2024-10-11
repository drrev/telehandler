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

// ValidateAccess that the given name has access to the given resource by checking
// if the name is [adminUser] or if the parent matches the given name.
func ValidateAccess(r Resource, name string) error {
	if adminUser != name && r.Parent() != name {
		return fmt.Errorf("user '%v' does not have permission to access resource '%v'", name, r.Identity())
	}
	return nil
}
