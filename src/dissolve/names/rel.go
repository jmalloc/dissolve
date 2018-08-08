package names

import (
	"errors"
	"fmt"
	"strings"
)

// Rel is a relative name.
//
// It differs from a hostname in that it MAY contain dots, but is not itself
// fully-qualified, and hence does not end in a trailing dot.
type Rel string

// ParseRel parses n as a relative name.
func ParseRel(n string) (Rel, error) {
	v := Rel(n)
	return v, v.Validate()
}

// MustParseRel parses n as a relative name.
// It panics if n is invalid.
func MustParseRel(n string) Rel {
	v, err := ParseRel(n)
	if err != nil {
		panic(err)
	}
	return v
}

// IsRelative returns true.
func (n Rel) IsRelative() bool {
	return true
}

// Qualify returns n, qualified against f.
func (n Rel) Qualify(f FQDN) FQDN {
	return MustParseFQDN(string(n) + "." + string(f))
}

// Split splits the "hostname" from the name.
// If the name does not contain any dots, tail is nil.
func (n Rel) Split() (head Host, tail Name) {
	s := string(n)
	i := strings.Index(s, ".")

	if i == -1 {
		head = MustParseHost(s)
	} else {
		head = MustParseHost(s[:i])
		tail = FQDN(s[i+1:])
	}

	return
}

// Validate returns nil if the name is valid.
func (n Rel) Validate() error {
	if n == "" {
		return errors.New("relative name must not be empty")
	}

	s := string(n)

	if strings.HasPrefix(s, ".") {
		return fmt.Errorf("relative name '%s' is invalid, unexpected leading dot", n)
	}

	if strings.HasSuffix(s, ".") {
		return fmt.Errorf("relative name '%s' is invalid, unexpected trailing dot", n)
	}

	return nil
}

// String returns a human-readable representation of the name.
func (n Rel) String() string {
	return string(n)
}
