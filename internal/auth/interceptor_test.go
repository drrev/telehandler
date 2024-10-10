package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"slices"
	"strings"
	"testing"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func TestTap(t *testing.T) {
	noError := func(e error) bool { return e == nil }
	errorTextContains := func(str string) func(e error) bool {
		return func(e error) bool {
			if e == nil {
				return len(str) == 0
			}
			return strings.Contains(e.Error(), str)
		}
	}

	tests := []struct {
		name    string
		ctx     context.Context
		names   []string
		wantErr func(e error) bool
	}{
		{
			name:    "invalid Context",
			ctx:     context.Background(),
			wantErr: errorTextContains("failed resolve peer"),
		},
		{
			name:    "invalid AuthInfo",
			ctx:     peer.NewContext(context.Background(), &peer.Peer{AuthInfo: nil}),
			wantErr: errorTextContains("invalid peer authentication"),
		},
		{
			name:    "empty CNs",
			ctx:     peer.NewContext(context.Background(), &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}}),
			wantErr: errorTextContains("no valid subject CN found"),
		},
		{
			name: "admin CN",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							{Subject: pkix.Name{CommonName: "admin"}},
						},
					},
				},
			}),
			names:   []string{"admin"},
			wantErr: noError,
		},
		{
			name: "multiple CNs",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							{Subject: pkix.Name{CommonName: "admin"}},
							{Subject: pkix.Name{CommonName: "test"}},
						},
					},
				},
			}),
			names:   []string{"admin", "test"},
			wantErr: noError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := Tap(tt.ctx, nil)
			if !tt.wantErr(err) {
				t.Errorf("ServerStreamInterceptor() error = %v", err)
			}

			if len(tt.names) > 0 {
				names, err := CommonNamesFromCtx(ctx)
				if err != nil {
					t.Errorf("CommonNamesFromCtx() error = %v", err)
				}

				if slices.Compare(names, tt.names) != 0 {
					t.Errorf("CommonNamesFromCtx() got %v, expected %v", names, tt.names)
				}
			}
		})
	}
}
