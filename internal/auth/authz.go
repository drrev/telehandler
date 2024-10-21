package auth

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

const adminUser = "admin"

func validateAccess(cn string, req any) error {
	// TODO: replace with OPA or OpenFGA or similar
	mpr := req.(proto.Message).ProtoReflect()
	parent := mpr.Descriptor().Fields().ByName("parent")
	name := mpr.Descriptor().Fields().ByName("name")

	switch {
	case parent != nil && mpr.Has(parent):
		val := mpr.Get(parent).String()
		pfx := fmt.Sprintf("users/%s", cn)
		if cn != adminUser && val != pfx {
			return status.Errorf(codes.PermissionDenied, "resource '%s' is not accessible by user '%s'", val, cn)
		}

	case name != nil && mpr.Has(name):
		val := mpr.Get(name).String()
		pfx := fmt.Sprintf("users/%s/", cn)
		if cn != adminUser && !strings.HasPrefix(val, pfx) {
			return status.Errorf(codes.PermissionDenied, "resource '%s' is not accessible by user '%s'", val, cn)
		}
	default:
		return status.Error(codes.PermissionDenied, "message has no 'parent' or 'name'")
	}

	return nil
}
