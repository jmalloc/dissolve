package names

import (
	"errors"
	"fmt"
	"strings"
)

// Host is the name of an internet host. Host names do not contain any dots.
type Host string

// ParseHost parses n as a host name.
func ParseHost(n string) (Host, error) {
	v := Host(n)
	return v, v.Validate()
}

// MustParseHost parses n as a Host.
// It panics if n is invalid.
func MustParseHost(n string) Host {
	v, err := ParseHost(n)
	if err != nil {
		panic(err)
	}
	return v
}

// IsRelative returns true.
func (n Host) IsRelative() bool {
	return true
}

// Qualify returns n, qualified against f.
func (n Host) Qualify(f FQDN) FQDN {
	return MustParseFQDN(string(n) + "." + string(f))
}

// Split splits the "hostname" from the name.
// If the name does not contain any dots, tail is nil.
func (n Host) Split() (head Host, tail Name) {
	head = n
	return
}

// Validate returns nil if the name is valid.
func (n Host) Validate() error {
	if n == "" {
		return errors.New("hostname must not be empty")
	}

	if strings.Contains(string(n), ".") {
		return fmt.Errorf("hostname '%s' is invalid, contains unexpected dots", n)
	}

	return nil
}

// String returns a human-readable representation of the name.
func (n Host) String() string {
	return string(n)
}
