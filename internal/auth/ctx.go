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

// resolveCommonName uses the gRPC request context to resolve the peer's certificates
// then resolves all non-empty Common Names. Returns a gRPC status error if no non-empty subject CNs were found.
func resolveCommonName(ctx context.Context) (cn string, err error) {
	// auth mTLS cert
	peer, ok := peer.FromContext(ctx)
	if !ok {
		err = status.Error(codes.Unauthenticated, "failed resolve peer")
		return
	}

	mtls, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		err = status.Errorf(codes.Unauthenticated, "invalid peer authentication")
		return
	}

	// this should be impossible, but bounds checking nonetheless
	if len(mtls.State.PeerCertificates) < 1 {
		err = status.Errorf(codes.Unauthenticated, "no certificates found")
		return
	}

	cn = mtls.State.PeerCertificates[0].Subject.CommonName

	if len(cn) < 1 {
		err = status.Error(codes.Unauthenticated, "no valid subject CN found")
	}

	return
}

// CommonNameToCtx adds cns to the given context.
// Use [CommonNameFromCtx] to get cns out of the context.
func CommonNameToCtx(ctx context.Context, cn string) context.Context {
	return context.WithValue(ctx, cnkey, cn)
}

// CommonNameFromCtx retrieves any Subject Common Names stored in the given context.
// If none are found, an error is returned.
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
