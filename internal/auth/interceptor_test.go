package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"slices"
	"strings"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func TestServerStreamInterceptor(t *testing.T) {
	mockHandler := func(expectedNames []string) grpc.StreamHandler {
		return func(srv any, stream grpc.ServerStream) error {
			names, err := CommonNamesFromCtx(stream.Context())
			if err != nil {
				return err
			}

			if slices.Compare(names, expectedNames) != 0 {
				return fmt.Errorf("invalid names (expected %#v, got %#v)", expectedNames, names)
			}

			return nil
		}
	}

	noError := func(e error) bool { return e == nil }
	errorTextContains := func(str string) func(e error) bool {
		return func(e error) bool {
			if e == nil {
				return len(str) == 0
			}
			return strings.Contains(e.Error(), str)
		}
	}

	type args struct {
		ss      grpc.ServerStream
		handler grpc.StreamHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr func(e error) bool
	}{
		{
			name:    "invalid Context",
			args:    args{ss: &mockServerStream{ctx: context.Background()}},
			wantErr: errorTextContains("failed resolve peer"),
		},
		{
			name:    "invalid AuthInfo",
			args:    args{ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{AuthInfo: nil})}},
			wantErr: errorTextContains("invalid peer authentication"),
		},
		{
			name:    "empty CNs",
			args:    args{ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}})}},
			wantErr: errorTextContains("no valid subject CN found"),
		},
		{
			name: "admin CN",
			args: args{
				handler: mockHandler([]string{"admin"}),
				ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{
					AuthInfo: credentials.TLSInfo{
						State: tls.ConnectionState{
							PeerCertificates: []*x509.Certificate{
								{Subject: pkix.Name{CommonName: "admin"}},
							},
						},
					},
				})},
			},
			wantErr: noError,
		},
		{
			name: "multiple CNs",
			args: args{
				handler: mockHandler([]string{"admin", "test"}),
				ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{
					AuthInfo: credentials.TLSInfo{
						State: tls.ConnectionState{
							PeerCertificates: []*x509.Certificate{
								{Subject: pkix.Name{CommonName: "admin"}},
								{Subject: pkix.Name{CommonName: "test"}},
							},
						},
					},
				})},
			},
			wantErr: noError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ServerStreamInterceptor(nil, tt.args.ss, nil, tt.args.handler); !tt.wantErr(err) {
				t.Errorf("ServerStreamInterceptor() error = %v", err)
			}
		})
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}
