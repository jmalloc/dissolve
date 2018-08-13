package names

import "strings"

// Name is an DNS name of some kind.
//
// Any of the methods except Validate() MAY panic if the name is invalid.
type Name interface {
	// IsQualified returns true if the name is fully-qualified.
	IsQualified() bool

	// Qualify returns a fully-qualified domain name produced by "qualifying"
	// this name with f.
	//
	// If this name is already qualified, it is returned unchanged.
	Qualify(f FQDN) FQDN

	// Labels returns the DNS labels that form this name.
	Labels() []Label

	// Split splits the first label from the name.
	// If the name only has single label, tail is nil.
	Split() (head Label, tail Name)

	// Join returns a name produced by concatenating this name with s.
	// It panics if this name is fully qualified.
	Join(s Name) Name

	// Validate returns nil if the name is valid.
	Validate() error

	// String returns a representation of the name as used by DNS systems.
	// It panics if the name is not valid.
	String() string
}

// Parse parses an arbitrary internet name.
func Parse(n string) (Name, error) {
	i := strings.Index(n, ".")

	var name Name

	if i == -1 {
		name = Label(n)
	} else if i == len(n)-1 {
		name = FQDN(n)
	} else {
		name = UDN(n)
	}

	return name, name.Validate()
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
