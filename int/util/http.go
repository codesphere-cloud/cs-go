package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// GetEphemeralPort finds an available ephemeral port on localhost.
func GetEphemeralPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer func() { _ = l.Close() }()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// WaitForPort blocks until a TCP port is listening or timeout occurs.
func WaitForPort(addr string, timeout time.Duration) {
	Eventually(func() error {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
		}
		return err
	}, timeout, "100ms").Should(Succeed(), fmt.Sprintf("Port %s should be listening", addr))
}

// Helper to start a simple HTTP server (for non-TLS tests)
func StartTestHttpServer(port int) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK (HTTP)"))
	})

	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{Addr: serverAddr, Handler: mux}

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", serverAddr, err)
	}

	go func() {
		defer GinkgoRecover() // Ensure goroutine panics are caught by Ginkgo
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			Fail(fmt.Sprintf("Test HTTP server failed: %v", err))
		}
	}()
	return server, nil
}

// Helper to start a simple HTTPS server with custom certs
func StartTestHttpsServer(port int, certFile, keyFile string) (*http.Server, error) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK (HTTPS)"))
	})

	serverAddr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{Addr: serverAddr, Handler: mux}

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", serverAddr, err)
	}

	go func() {
		defer GinkgoRecover() // Ensure goroutine panics are caught by Ginkgo
		if err := server.ServeTLS(listener, certFile, keyFile); err != nil && err != http.ErrServerClosed {
			Fail(fmt.Sprintf("Test HTTPS server failed: %v", err))
		}
	}()
	return server, nil
}

// GenerateTLSCerts generates a CA, a server key, and a server certificate signed by the CA.
// It includes Subject Alternative Names (SANs) for modern TLS compliance.
// Returns paths to ca.crt, server.crt, and server.key.
func GenerateTLSCerts(outputDir string, serverCN string, serverSANs []string) (caCertPath, serverCertPath, serverKeyPath string, err error) {
	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", "", "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// CA key and cert
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate CA private key: %w", err)
	}
	caTemplate := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Testcontainers CA"},
			CommonName:   "CA Root for " + serverCN,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(3650 * 24 * time.Hour), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, &caTemplate, &caTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create CA certificate: %w", err)
	}

	caCertPath = filepath.Join(outputDir, "ca.crt")
	caKeyPath := filepath.Join(outputDir, "ca.key")
	if err := os.WriteFile(caCertPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write CA certificate: %w", err)
	}
	if err := os.WriteFile(caKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caKey)}), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to write CA key: %w", err)
	}

	// Server key and cert
	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to generate server private key: %w", err)
	}
	serverTemplate := x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject: pkix.Name{
			Organization: []string{"Test App Service"},
			CommonName:   serverCN,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 year
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Add SANs
	for _, san := range serverSANs {
		if ip := net.ParseIP(san); ip != nil {
			serverTemplate.IPAddresses = append(serverTemplate.IPAddresses, ip)
		} else {
			serverTemplate.DNSNames = append(serverTemplate.DNSNames, san)
		}
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, &serverTemplate, &caTemplate, &serverKey.PublicKey, caKey)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create server certificate: %w", err)
	}

	serverCertPath = filepath.Join(outputDir, "server.crt")
	serverKeyPath = filepath.Join(outputDir, "server.key")
	if err := os.WriteFile(serverCertPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}), 0644); err != nil {
		return "", "", "", fmt.Errorf("failed to write server certificate: %w", err)
	}
	if err := os.WriteFile(serverKeyPath, pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)}), 0600); err != nil {
		return "", "", "", fmt.Errorf("failed to write server key: %w", err)
	}

	return caCertPath, serverCertPath, serverKeyPath, nil
}
