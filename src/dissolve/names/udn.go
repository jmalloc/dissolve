package names

import (
	"errors"
	"fmt"
	"strings"
)

// UDN is a unqualified (relative) domain name that may multiple DNS labels.
type UDN string

// IsQualified returns false.
func (n UDN) IsQualified() bool {
	return false
}

// Qualify returns a fully-qualified domain name produced by "qualifying"
// this name with f.
func (n UDN) Qualify(f FQDN) FQDN {
	return FQDN(n.String() + "." + f.String())
}

// Labels returns the DNS labels that form this name.
// It panics if the name is not valid.
func (n UDN) Labels() []Label {
	s := n.String()
	var labels []Label

	for {
		i := strings.Index(s, ".")
		if i == -1 {
			return append(labels, Label(s))
		}

		labels = append(labels, Label(s[:i]))
		s = s[i+1:]
	}
}

// Split splits the first label from the name.
// If the name only has single label, tail is nil.
func (n UDN) Split() (head Label, tail Name) {
	s := n.String()
	i := strings.Index(s, ".")

	head = Label(s[:i])

	if i != -1 {
		tail = UDN(s[i:])
	}

	return
}

// Join returns a name produced by concatenating this name with s.
// It panics if this name is fully qualified.
func (n UDN) Join(s Name) Name {
	return MustParse(n.String() + "." + s.String())
}

// Validate returns nil if the name is valid.
func (n UDN) Validate() error {
	if n == "" {
		return errors.New("unqualified domain name must not be empty")
	}

	if n[0] == '.' {
		return fmt.Errorf("unqualified domain name '%s' is invalid, unexpected leading dot", n)
	}

	if n[len(n)-1] == '.' {
		return fmt.Errorf("unqualified domain name '%s' is invalid, unexpected trailing dot", n)
	}

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n UDN) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	return string(n)
}
