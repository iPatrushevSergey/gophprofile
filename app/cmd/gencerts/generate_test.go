package main

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate_writesValidDevCerts(t *testing.T) {
	dir := t.TempDir()

	err := Generate(Options{
		OutDir:     dir,
		ServerName: "localhost",
		Hosts:      []string{"localhost", "127.0.0.1"},
		Days:       30,
		CAName:     "Test CA",
		Force:      false,
	})
	require.NoError(t, err)

	paths := newCertPaths(dir)
	caCert := readCertificate(t, paths.caCert)
	serverCert := readCertificate(t, paths.serverCert)

	assert.True(t, caCert.IsCA)
	assert.Contains(t, serverCert.DNSNames, "localhost")
	assert.Len(t, serverCert.IPAddresses, 1)

	roots := x509.NewCertPool()
	roots.AddCert(caCert)
	_, err = serverCert.Verify(x509.VerifyOptions{
		DNSName: "localhost",
		Roots:   roots,
	})
	require.NoError(t, err)
}

func TestGenerate_refusesOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	opts := Options{
		OutDir:     dir,
		ServerName: "localhost",
		Hosts:      []string{"localhost"},
		Days:       30,
		CAName:     "Test CA",
	}

	require.NoError(t, Generate(opts))
	err := Generate(opts)
	require.Error(t, err)
}

func readCertificate(t *testing.T, path string) *x509.Certificate {
	t.Helper()

	raw, err := os.ReadFile(filepath.Clean(path))
	require.NoError(t, err)

	block, _ := pem.Decode(raw)
	require.NotNil(t, block)

	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return cert
}
