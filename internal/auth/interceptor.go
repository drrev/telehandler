package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/tap"
)

// Tap creates a [tap.ServerInHandle] to intercept connections, collect mTLS CNs, and populate the context
// passed into all handlers. So, each handler will receive a [context.Context] with CNs populated.
//
// Use [CommonNameFromCtx] to access CNs from the context.
func Tap(ctx context.Context, _ *tap.Info) (context.Context, error) {
	cns, err := resolveCommonName(ctx)
	if err != nil {
		return ctx, err
	}

	return CommonNameToCtx(ctx, cns), nil
}

// UnaryServerInterceptor enforces basic resource authorization.
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	cn, err := CommonNameFromCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "missing valid common name")
	}

	if err := validateAccess(cn, req); err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

// ServerStreamInterceptor enforces basic resource authorization.
func ServerStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, &serverStream{ServerStream: ss})
}

// serverStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type serverStream struct {
	grpc.ServerStream
}

func (w *serverStream) RecvMsg(m interface{}) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	cn, err := CommonNameFromCtx(w.Context())
	if err != nil {
		return status.Error(codes.PermissionDenied, "missing valid common name")
	}

	if err := validateAccess(cn, m); err != nil {
		return err
	}
	return nil
}
