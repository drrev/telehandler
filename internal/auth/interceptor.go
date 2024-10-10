package auth

import (
	"context"

	"google.golang.org/grpc/tap"
)

// Tap creates a [tap.ServerInHandle] to intercept connections, collect mTLS CNs, and populate the context
// passed into all handlers. So, each handler will receive a [context.Context] with CNs populated.
//
// Use [CommonNamesFromCtx] to access CNs from the context.
func Tap(ctx context.Context, _ *tap.Info) (context.Context, error) {
	cns, err := resolveCommonNames(ctx)
	if err != nil {
		return ctx, err
	}

	return CommonNamesToCtx(ctx, cns), nil
}
