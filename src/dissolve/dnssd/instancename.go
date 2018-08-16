package dnssd

import (
	"errors"
	"strings"

	"github.com/jmalloc/dissolve/src/dissolve/names"
)

// InstanceName is an implementation of names.Name that represents an
// unqualified instance name.
type InstanceName string

// SplitInstanceName parses the first label of n as a backslash-escaped instance
// name. If n contains only a single label, tail is nil.
func SplitInstanceName(n names.Name) (head InstanceName, tail names.Name) {
	s := n.String()

	var b strings.Builder
	b.Grow(len(s))

	// https://tools.ietf.org/html/rfc6763#section-4.3
	//
	// This document RECOMMENDS that if concatenating the three portions of
	// a Service Instance Name, any dots in the <Instance> portion be
	// escaped following the customary DNS convention for text files: by
	// preceding literal dots with a backslash (so "." becomes "\.").
	// Likewise, any backslashes in the <Instance> portion should also be
	// escaped by preceding them with a backslash (so "\" becomes "\\").
	esc := false

	for i := 0; i < len(s); i++ {
		c := s[i]

		if esc {
			// accept any character after a backslash
			b.WriteByte(c)
			esc = false
		} else if c == '\\' {
			esc = true
		} else if c == '.' {
			head = InstanceName(b.String())
			if i < len(s)-1 {
				tail = names.MustParse(s[i+1:])
			}
			return
		} else {
			b.WriteByte(c)
		}
	}

	// if the name ends midway through an escape sequence, we assume the string was
	// intended to end with a backslash.
	if esc {
		b.WriteByte('\\')
	}

	head = InstanceName(b.String())
	return
}

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

	return nil
}

// String returns a representation of the name as used by DNS systems.
// It panics if the name is not valid.
func (n InstanceName) String() string {
	if err := n.Validate(); err != nil {
		panic(err)
	}

	s := string(n)
	var b strings.Builder
	b.Grow(len(s) * 2)

	for {
		i := strings.IndexAny(s, `.\`)
		if i == -1 {
			b.WriteString(s)
			break
		}

		b.WriteString(s[:i])
		b.WriteByte('\\')
		b.WriteByte(s[i])
		s = s[i+1:]
	}

	return b.String()
}
