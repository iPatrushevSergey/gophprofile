package config

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Address is a custom type for host:port.
type Address struct {
	Schema string
	Host   string
	Port   int
}

// Set implements flag.Value.
func (a *Address) Set(s string) error {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = "http://" + s
	}

	u, err := url.Parse(s)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}

	if u.Host == "" {
		return errors.New("host is empty")
	}

	hostName, portStr, err := net.SplitHostPort(u.Host)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}

	a.Schema = u.Scheme
	a.Host = hostName
	a.Port = port

	return nil
}

// String returns host:port listen address.
func (a *Address) String() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}

// URL returns full base URL.
func (a *Address) URL() string {
	return fmt.Sprintf("%s://%s:%d", a.Schema, a.Host, a.Port)
}

// UnmarshalText supports env parsing.
func (a *Address) UnmarshalText(text []byte) error {
	return a.Set(string(text))
}

// Duration wraps time.Duration for config parsing.
type Duration struct {
	time.Duration
}

// Set implements flag.Value.
func (d *Duration) Set(s string) error {
	if val, err := strconv.Atoi(s); err == nil {
		d.Duration = time.Duration(val) * time.Second
		return nil
	}
	val, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = val
	return nil
}

// String returns duration string.
func (d *Duration) String() string {
	return d.Duration.String()
}

// UnmarshalText supports env parsing.
func (d *Duration) UnmarshalText(text []byte) error {
	return d.Set(string(text))
}
