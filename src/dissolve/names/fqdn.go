package names

import (
	"errors"
	"fmt"
	"strings"
)

// FQDN is a fully-qualified internet domain name.
type FQDN string

// IsQualified returns true.
func (n FQDN) IsQualified() bool {
	return true
}

// Qualify returns n unchanged.
func (n FQDN) Qualify(FQDN) FQDN {
	return n
}

// Labels returns the DNS labels that form this name.
// It panics if the name is not valid.
func (n FQDN) Labels() []Label {
	s := n.String()
	var labels []Label

	for {
		i := strings.Index(s, ".")
		if i == -1 {
			return labels
		}

		labels = append(labels, Label(s[:i]))
		s = s[i+1:]
	}
}

// Split splits the first label from the name.
// If the name only has single label, tail is nil.
// It panics if the name is not valid.
func (n FQDN) Split() (head Label, tail Name) {
	s := n.String()
	i := strings.Index(s, ".")

	head = Label(s[:i])

	if i != len(s)-1 {
		tail = FQDN(s[i:])
	}

	return
}

// Join returns a name produced by concatenating this name with s.
// It panics if this name is fully qualified.
func (n FQDN) Join(s Name) Name {
	panic(fmt.Sprintf(
		"can not join '%s' to '%s', left-hand-side is already fully-qualified",
		n,
		s,
	))
}

// Validate returns nil if the name is valid.
func (n FQDN) Validate() error {
	if n == "" {
		return errors.New("fully-qualified name must not be empty")
	}

	if n[0] == '.' {
		return fmt.Errorf("fully-qualified name '%s' is invalid, unexpected leading dot", n)
	}

	if n[len(n)-1] != '.' {
		return fmt.Errorf("fully-qualified name '%s' is invalid, missing trailing dot", n)
	}

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n FQDN) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	return string(n)
}
