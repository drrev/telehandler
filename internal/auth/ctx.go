package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type cnctxkey struct{}

var cnkey = cnctxkey{}

// resolveCommonNames uses the gRPC request context to resolve the peer's certificates
// then resolves all non-empty Common Names. Returns a gRPC status error if no non-empty subject CNs were found.
func resolveCommonNames(ctx context.Context) ([]string, error) {
	// auth mTLS cert
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "failed resolve peer")
	}

	mtls, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "invalid peer authentication")
	}

	// pull out all common names
	names := []string{}
	for _, item := range mtls.State.PeerCertificates {
		if len(item.Subject.CommonName) > 0 {
			names = append(names, item.Subject.CommonName)
		}
	}

	if len(names) < 1 {
		return nil, status.Error(codes.Unauthenticated, "no valid subject CN found")
	}

	return names, nil
}

// CommonNamesToCtx adds cns to the given context.
// Use [CommonNamesFromCtx] to get cns out of the context.
func CommonNamesToCtx(ctx context.Context, cns []string) context.Context {
	return context.WithValue(ctx, cnkey, cns)
}

// CommonNamesFromCtx retrieves any Subject Common Names stored in the given context.
// If none are found, an error is returned.
func CommonNamesFromCtx(ctx context.Context) ([]string, error) {
	v := ctx.Value(cnkey)
	if v == nil {
		return nil, fmt.Errorf("missing CommonName")
	}

	cns, ok := v.([]string)
	if !ok {
		return nil, fmt.Errorf("invalid CommonName type: '%T'", v)
	}

	return cns, nil
}
