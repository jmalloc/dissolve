package dnssd

import (
	"errors"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// InstanceName is an implementation of names.Name that represents an
// unqualified instance name.
type InstanceName string

// IsQualified returns false.
func (n InstanceName) IsQualified() bool {
	return false
}

// Qualify returns a fully-qualified domain name produced by "qualifying"
// this name with f.
func (n InstanceName) Qualify(f names.FQDN) names.FQDN {
	return names.FQDN(n.String() + "." + f.String())
}

// Labels returns the DNS labels that form this name.
func (n InstanceName) Labels() []names.Label {
	h, _ := n.Split()
	return []names.Label{h}
}

// Split splits the "hostname" from the name.
// If the name does not contain any dots, tail is nil.
func (n InstanceName) Split() (head names.Label, tail names.Name) {
	head = names.Label(n.String())
	return
}

// Join returns a name produced by concatenating this name with s.
// It panics if this name is fully qualified.
func (n InstanceName) Join(s names.Name) names.Name {
	return names.MustParse(n.String() + "." + s.String())
}

// Validate returns nil if the name is valid.
func (n InstanceName) Validate() error {
	if n == "" {
		return errors.New("instance name must not be empty")
	}

	// TODO(jmalloc): actually validate

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n InstanceName) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	// TODO(jmalloc): escape

	return string(n)
}
