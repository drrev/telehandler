package auth

import (
	"context"

	"google.golang.org/grpc"
)

// UnaryServerInterceptor extracts CNs from peer certs and passes them to downstream handlers for authz.
func UnaryServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// If we were using resource naming for requests, it would be possible to generically authenticate requests.
	// Since we aren't, auth will have to be performed within each handler individually.
	// So, just resolve common names and inject them into ctx. Each handler will have to query any job and enforce authz.

	cns, err := resolveCommonNames(ctx)
	if err != nil {
		return nil, err
	}

	return handler(CommonNamesToCtx(ctx, cns), req)
}

// ServerStreamInterceptor extracts CNs from peer certs and passes them to downstream handlers for authz.
func ServerStreamInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()
	cns, err := resolveCommonNames(ctx)
	if err != nil {
		return err
	}
	return handler(srv, &wrappedStream{ServerStream: ss, ctx: CommonNamesToCtx(ctx, cns)})
}
