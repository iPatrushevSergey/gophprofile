package apputil

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildNA(t *testing.T) {
	assert.Equal(t, "N/A", buildNA(""))
	assert.Equal(t, "dev", buildNA("dev"))
}

func TestPrintBuildInfo(t *testing.T) {
	Version = "1.0.0"
	Date = "2026-01-01"

	var buf bytes.Buffer
	PrintBuildInfo(&buf)
	assert.Contains(t, buf.String(), "version: 1.0.0")
	assert.Contains(t, buf.String(), "date: 2026-01-01")
}

func TestHandleVersionArg(t *testing.T) {
	assert.False(t, HandleVersionArg([]string{"login"}))
	assert.True(t, HandleVersionArg([]string{"--version"}))
}
