package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
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
		cn      string
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
			wantErr: errorTextContains("Unauthenticated"),
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
			cn:      "admin",
			wantErr: noError,
		},
		{
			name: "multiple CNs",
			ctx: peer.NewContext(context.Background(), &peer.Peer{
				AuthInfo: credentials.TLSInfo{
					State: tls.ConnectionState{
						PeerCertificates: []*x509.Certificate{
							{Subject: pkix.Name{CommonName: "test"}},
							{Subject: pkix.Name{CommonName: "admin"}},
						},
					},
				},
			}),
			cn:      "test",
			wantErr: noError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := Tap(tt.ctx, nil)
			if !tt.wantErr(err) {
				t.Errorf("ServerStreamInterceptor() error = %v", err)
			}

			if len(tt.cn) > 0 {
				name, err := CommonNameFromCtx(ctx)
				if err != nil {
					t.Errorf("CommonNameFromCtx() error = %v", err)
				}

				if name != tt.cn {
					t.Errorf("CommonNameFromCtx() got %v, expected %v", name, tt.cn)
				}
			}
		})
	}
}
