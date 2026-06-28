package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// defaultKeyBits is the default key size for RSA keys.
const defaultKeyBits = 4096

// Options configures dev TLS certificate generation.
type Options struct {
	OutDir     string
	ServerName string
	Hosts      []string
	Days       int
	CAName     string
	Force      bool
}

// Generate writes a dev CA and a server certificate signed by that CA.
func Generate(opts Options) error {
	if strings.TrimSpace(opts.OutDir) == "" {
		return errors.New("output directory is required")
	}
	if strings.TrimSpace(opts.ServerName) == "" {
		return errors.New("server name is required")
	}
	if strings.TrimSpace(opts.CAName) == "" {
		return errors.New("ca name is required")
	}
	if opts.Days <= 0 {
		return errors.New("days must be positive")
	}
	if len(opts.Hosts) == 0 {
		return errors.New("at least one host is required")
	}

	outDir := filepath.Clean(opts.OutDir)
	paths := newCertPaths(outDir)
	if !opts.Force {
		for _, path := range paths.all() {
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("%s already exists (use -force to overwrite)", path)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("stat %s: %w", path, err)
			}
		}
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	caKey, err := rsa.GenerateKey(rand.Reader, defaultKeyBits)
	if err != nil {
		return fmt.Errorf("generate ca key: %w", err)
	}

	now := time.Now().UTC()
	caTmpl := &x509.Certificate{
		SerialNumber:          randomSerial(),
		Subject:               pkix.Name{CommonName: opts.CAName, Organization: []string{opts.CAName}},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(time.Duration(opts.Days) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create ca certificate: %w", err)
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, defaultKeyBits)
	if err != nil {
		return fmt.Errorf("generate server key: %w", err)
	}

	dnsNames, ipAddrs, err := parseHosts(opts.Hosts)
	if err != nil {
		return err
	}

	serverTmpl := &x509.Certificate{
		SerialNumber: randomSerial(),
		Subject: pkix.Name{
			CommonName:   opts.ServerName,
			Organization: []string{opts.CAName},
		},
		NotBefore:   now.Add(-time.Hour),
		NotAfter:    now.Add(time.Duration(opts.Days) * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:    dnsNames,
		IPAddresses: ipAddrs,
	}

	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create server certificate: %w", err)
	}

	if err := writePEM(paths.caCert, "CERTIFICATE", caDER, 0o644); err != nil {
		return err
	}
	if err := writePEM(paths.caKey, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(caKey), 0o600); err != nil {
		return err
	}
	if err := writePEM(paths.serverCert, "CERTIFICATE", serverDER, 0o644); err != nil {
		return err
	}
	if err := writePEM(paths.serverKey, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey), 0o600); err != nil {
		return err
	}

	return nil
}

// certPaths represents the paths to the certificate files.
type certPaths struct {
	caCert     string
	caKey      string
	serverCert string
	serverKey  string
}

// newCertPaths returns a new certPaths instance.
func newCertPaths(outDir string) certPaths {
	return certPaths{
		caCert:     filepath.Join(outDir, "ca.pem"),
		caKey:      filepath.Join(outDir, "ca-key.pem"),
		serverCert: filepath.Join(outDir, "server.crt"),
		serverKey:  filepath.Join(outDir, "server.key"),
	}
}

// all returns all certificate paths.
func (p certPaths) all() []string {
	return []string{p.caCert, p.caKey, p.serverCert, p.serverKey}
}

// parseHosts parses a list of hosts into DNS names and IP addresses.
func parseHosts(hosts []string) (dnsNames []string, ipAddrs []net.IP, err error) {
	for _, host := range hosts {
		host = strings.TrimSpace(host)
		if host == "" {
			continue
		}
		if ip := net.ParseIP(host); ip != nil {
			ipAddrs = append(ipAddrs, ip)
			continue
		}
		dnsNames = append(dnsNames, host)
	}
	if len(dnsNames) == 0 && len(ipAddrs) == 0 {
		return nil, nil, errors.New("no valid hosts")
	}
	return dnsNames, ipAddrs, nil
}

// randomSerial generates a random serial number for a certificate.
func randomSerial() *big.Int {
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return big.NewInt(time.Now().UnixNano())
	}
	return serial
}

// writePEM writes a PEM-encoded certificate or private key to a file.
func writePEM(path, blockType string, der []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	if err := pem.Encode(file, &pem.Block{Type: blockType, Bytes: der}); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
