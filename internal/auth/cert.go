package auth

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

// LoadServerTLS is a helper to create mTLS transport credentials and a cert pool
// using a self-signed cert chain for a gRPC server.
func LoadServerTLS(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load key pair: %w", err)
	}

	data, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    capool,
		MinVersion:   tls.VersionTLS13,
	}

	return credentials.NewTLS(tlsConfig), nil
}

// LoadServerTLS is a helper to create mTLS transport credentials and a cert pool
// using a self-signed cert chain for a gRPC client.
func LoadClientTLS(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certification: %w", err)
	}

	ca, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("faild to append the CA certificate to CA pool")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
	}

	return credentials.NewTLS(tlsConfig), nil
}
