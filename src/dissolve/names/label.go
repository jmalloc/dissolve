package names

import (
	"errors"
	"fmt"
	"strings"
)

// Label is the part of a DNS name contains within dots.
type Label string

// IsQualified returns false.
func (n Label) IsQualified() bool {
	return false
}

// Qualify returns a fully-qualified domain name produced by "qualifying"
// this name with f.
func (n Label) Qualify(f FQDN) FQDN {
	return FQDN(n.String() + "." + f.String())
}

// Labels returns the DNS labels that form this name.
func (n Label) Labels() []Label {
	return []Label{n}
}

// Split splits the "hostname" from the name.
// If the name does not contain any dots, tail is nil.
func (n Label) Split() (head Label, tail Name) {
	head = n
	return
}

// Join returns a name produced by concatenating this name with s.
// It panics if this name is fully qualified.
func (n Label) Join(s Name) Name {
	return MustParse(n.String() + "." + s.String())
}

// Validate returns nil if the name is valid.
func (n Label) Validate() error {
	if n == "" {
		return errors.New("label must not be empty")
	}

	if strings.Contains(string(n), ".") {
		return fmt.Errorf("label '%s' is invalid, contains unexpected dots", n)
	}

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n Label) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	return string(n)
}
