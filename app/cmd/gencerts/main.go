// Command gencerts generates a dev TLS CA and server certificate for local HTTPS.
package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// main is the entry point for the gencerts command.
func main() {
	outDir := flag.String("out", "certs", "output directory")
	serverName := flag.String("server-name", "localhost", "server hostname in certificate subject (CN)")
	hosts := flag.String("hosts", "localhost,127.0.0.1", "comma-separated DNS names and IPs for SAN")
	days := flag.Int("days", 365, "certificate validity in days")
	caName := flag.String("ca-name", "GophProfile Dev CA", "display name of the dev certificate authority")
	force := flag.Bool("force", false, "overwrite existing files")
	flag.Parse()

	opts := Options{
		OutDir:     *outDir,
		ServerName: *serverName,
		Hosts:      splitHosts(*hosts),
		Days:       *days,
		CAName:     *caName,
		Force:      *force,
	}

	if err := Generate(opts); err != nil {
		log.Fatalf("gencerts: %v", err)
	}

	paths := newCertPaths(opts.OutDir)
	fmt.Printf("wrote dev TLS certificates in %s:\n", opts.OutDir)
	fmt.Printf("  %s  (optional trust store for browser/curl)\n", paths.caCert)
	fmt.Printf("  %s  (server cert_file)\n", paths.serverCert)
	fmt.Printf("  %s  (server key_file)\n", paths.serverKey)
	fmt.Printf("  %s  (dev CA private key, keep secret)\n", paths.caKey)
}

// splitHosts splits a comma-separated list of hosts into a slice of strings.
func splitHosts(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
