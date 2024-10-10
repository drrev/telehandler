package auth

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"fmt"
	"slices"
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

	type args struct {
		ss      grpc.ServerStream
		handler grpc.StreamHandler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "invalid Context",
			args:    args{ss: &mockServerStream{ctx: context.Background()}},
			wantErr: true,
		},
		{
			name:    "invalid AuthInfo",
			args:    args{ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{AuthInfo: nil})}},
			wantErr: true,
		},
		{
			name:    "empty CNs",
			args:    args{ss: &mockServerStream{ctx: peer.NewContext(context.Background(), &peer.Peer{AuthInfo: credentials.TLSInfo{State: tls.ConnectionState{}}})}},
			wantErr: true,
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
			wantErr: false,
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
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ServerStreamInterceptor(nil, tt.args.ss, nil, tt.args.handler); (err != nil) != tt.wantErr {
				t.Errorf("ServerStreamInterceptor() error = %v, wantErr %v", err, tt.wantErr)
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
