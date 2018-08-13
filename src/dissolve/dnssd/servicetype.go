package dnssd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// ServiceType is an implementation of names.Name that represents a DNS-SD
// service type.
type ServiceType string

// IsQualified returns false.
func (n ServiceType) IsQualified() bool {
	return false
}

// Qualify returns a fully-qualified domain name produced by "qualifying"
// this name with f.
func (n ServiceType) Qualify(f names.FQDN) names.FQDN {
	return names.FQDN(n.String() + "." + f.String())
}

// Labels returns the DNS labels that form this name.
// It panics if the name is not valid.
func (n ServiceType) Labels() []names.Label {
	s := n.String()
	var labels []names.Label

	for {
		i := strings.Index(s, ".")
		if i == -1 {
			return append(labels, names.Label(s))
		}

		labels = append(labels, names.Label(s[:i]))
		s = s[i+1:]
	}
}

// Split splits the first label from the name.
// If the name only has single label, tail is nil.
func (n ServiceType) Split() (head names.Label, tail names.Name) {
	s := n.String()
	i := strings.Index(s, ".")

	head = names.Label(s[:i])

	if i != -1 {
		tail = names.UDN(s[i:])
	}

	return
}

// Join returns a name produced by concatenating this name with s.
// It panics if this name is fully qualified.
func (n ServiceType) Join(s names.Name) names.Name {
	return names.MustParse(n.String() + "." + s.String())
}

// Validate returns nil if the name is valid.
func (n ServiceType) Validate() error {
	if n == "" {
		return errors.New("service type must not be empty")
	}

	if n[0] == '.' {
		return fmt.Errorf("service type '%s' is invalid, unexpected leading dot", n)
	}

	if n[len(n)-1] == '.' {
		return fmt.Errorf("service type '%s' is invalid, unexpected trailing dot", n)
	}

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n ServiceType) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	return string(n)
}
