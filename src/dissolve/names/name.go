package names

import "strings"

// Name is an internet name of some kind.
type Name interface {
	// IsRelative returns true if the name is relative.
	IsRelative() bool

	// Qualify returns a fully-qualified domain name produced by "qualifying"
	// this name with f.
	//
	// If this name is already qualified, it is returned unchanged.
	Qualify(f FQDN) FQDN

	// Split splits the "hostname" from the name.
	// If the name does not contain any dots, tail is nil.
	Split() (head Host, tail Name)

	// Validate returns nil if the name is valid.
	Validate() error

	// String returns a human-readable string representation of the name.
	String() string
}

// Parse parses an arbitrary internet name.
func Parse(n string) (Name, error) {
	i := strings.Index(n, ".")

	if i == -1 {
		return ParseHost(n)
	} else if i == len(n)-1 {
		return ParseFQDN(n)
	}

	return ParseRel(n)
}

// MustParse parses an arbitrary internet name.
// It panics if n is invalid.
func MustParse(n string) Name {
	v, err := Parse(n)
	if err != nil {
		panic(err)
	}
	return v
}
