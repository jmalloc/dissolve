package names

import (
	"errors"
	"fmt"
	"strings"
)

// FQDN is a fully-qualified internet domain name.
type FQDN string

// ParseFQDN parses n as a FQDN.
func ParseFQDN(n string) (FQDN, error) {
	v := FQDN(n)
	return v, v.Validate()
}

// MustParseFQDN parses n as a FQDN.
// It panics if n is invalid.
func MustParseFQDN(n string) FQDN {
	v, err := ParseFQDN(n)
	if err != nil {
		panic(err)
	}
	return v
}

// IsRelative returns false.
func (n FQDN) IsRelative() bool {
	return false
}

// Qualify returns n unchanged.
func (n FQDN) Qualify(FQDN) FQDN {
	return n
}

// Split splits the "hostname" from the name.
// If the name does not contain any dots, tail is nil.
func (n FQDN) Split() (head Host, tail Name) {
	s := string(n)
	i := strings.Index(s, ".")

	if i == -1 {
		panic(n.Validate())
	}

	if i != len(n)-1 {
		tail = FQDN(s[i+1:])
	}

	head = MustParseHost(s[:i])
	return
}

// Validate returns nil if the name is valid.
func (n FQDN) Validate() error {
	if n == "" {
		return errors.New("fully-qualified name must not be empty")
	}

	s := string(n)

	if strings.HasPrefix(s, ".") {
		return fmt.Errorf("fully-qualified name '%s' is invalid, unexpected leading dot", n)
	}

	if !strings.HasSuffix(s, ".") {
		return fmt.Errorf("fully-qualified name '%s' is invalid, missing trailing dot", n)
	}

	return nil
}

// String returns a human-readable representation of the name.
func (n FQDN) String() string {
	return string(n)
}

// DNSString returns the string representation of the FQDN for use with DNS systems.
func (n FQDN) DNSString() string {
	return string(n)
}

// CommonString returns the "common" string representation of the FQDN, that is
// without the trailing dot.
func (n FQDN) CommonString() string {
	return strings.TrimSuffix(string(n), ".")
}
