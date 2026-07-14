package apputil

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Set at link time via -ldflags:
// -X github.com/iPatrushevSergey/gophprofile/app/internal/pkg/apputil.Version=...
var (
	Version string
	Date    string
)

// HandleVersionArg prints build metadata and reports whether the process should exit.
func HandleVersionArg(args []string) bool {
	for _, arg := range args {
		if arg == "--version" || arg == "-V" {
			PrintBuildInfo(os.Stdout)
			return true
		}
	}
	return false
}

// PrintBuildInfo writes build metadata to w.
func PrintBuildInfo(w io.Writer) {
	_, _ = fmt.Fprintf(w, "version: %s\n", buildNA(Version))
	_, _ = fmt.Fprintf(w, "date: %s\n", buildNA(Date))
}

func buildNA(s string) string {
	if strings.TrimSpace(s) == "" {
		return "N/A"
	}
	return s
}
