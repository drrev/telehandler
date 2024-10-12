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

// resolveCommonName uses the gRPC request context to resolve the peer's Common Name.
// Returns a gRPC status error if no subject CN was found.
func resolveCommonName(ctx context.Context) (string, error) {
	// auth mTLS cert
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "failed resolve peer")
	}

	mtls, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return "", status.Errorf(codes.Unauthenticated, "invalid peer authentication")
	}

	// this should be impossible, but bounds checking nonetheless
	if len(mtls.State.PeerCertificates) < 1 {
		return "", status.Errorf(codes.Unauthenticated, "no certificates found")
	}

	cn := mtls.State.PeerCertificates[0].Subject.CommonName

	if len(cn) < 1 {
		return cn, status.Error(codes.Unauthenticated, "no valid subject CN found")
	}

	return cn, nil
}

// CommonNameToCtx adds a CN to the given context.
// Use [CommonNameFromCtx] to get the CN out of the context.
func CommonNameToCtx(ctx context.Context, cn string) context.Context {
	return context.WithValue(ctx, cnkey, cn)
}

// CommonNameFromCtx retrieves any CN stored in the given context.
// If one is not found, an error is returned.
func CommonNameFromCtx(ctx context.Context) (string, error) {
	v := ctx.Value(cnkey)
	if v == nil {
		return "", fmt.Errorf("missing CommonName")
	}

	cns, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("invalid CommonName type: '%T'", v)
	}

	return cns, nil
}
