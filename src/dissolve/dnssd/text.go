package dnssd

// Text is a map that represents the key/value pairs in
// a service instance's TXT record.
//
// Keys are case-insensitive. The specification states that keys SHOULD be no
// longer than 9 characters. However since this is not a strict requirement, no
// such limit is enforced by this implementation.
//
// See https://tools.ietf.org/html/rfc6763#section-6.1
type Text struct {
	m map[string]string
}

// Get returns the first value that is associated with the key k.
func (t *Text) Get(k string) (string, bool) {
	if err := ValidateTextKey(k); err != nil {
		panic(err)
	}

	v, ok := t.m[k]
	return v, ok
}

// Set associates the value v with the key k.
// It is recommended that keys be no longer than 9 characters.
func (t *Text) Set(k string, v string) {
	if err := ValidateTextKey(k); err != nil {
		panic(err)
	}

	if err := ValidateTextValue(v); err != nil {
		panic(err)
	}

	if t.m == nil {
		t.m = map[string]string{k: v}
	} else {
		t.m[k] = v
	}
}

// SetBool associates an empty value with k if v is true; otherwise, it deletes
// the value associated with k, if any.
func (t *Text) SetBool(k string, v bool) {
	if v {
		t.Set(k, "")
	} else {
		t.Delete(k)
	}
}

// GetBool returns true if k is present in the map.
//
// This method is similar to Has(). It is included to better express intent when
// a key is used as a boolean value as per the recommendations in
// https://tools.ietf.org/html/rfc6763#section-6.4.
func (t *Text) GetBool(k string) bool {
	if err := ValidateTextKey(k); err != nil {
		panic(err)
	}

	_, ok := t.m[k]
	return ok
}

// Has returns true if all of the keys in k are present in the map.
func (t *Text) Has(k ...string) bool {
	for _, x := range k {
		if err := ValidateTextKey(x); err != nil {
			panic(err)
		}

		if _, ok := t.m[x]; !ok {
			return false
		}
	}

	return true
}

// Delete removes all of the given keys from the map.
func (t *Text) Delete(k ...string) {
	for _, x := range k {
		if err := ValidateTextKey(x); err != nil {
			panic(err)
		}

		delete(t.m, x)
	}
}

// Pairs returns the string representation of each key/value pair, as they appear
// in the TXT record.
func (t *Text) Pairs() []string {
	pairs := make([]string, 0, len(t.m))

	for k, v := range t.m {
		if v == "" {
			pairs = append(pairs, k)
		} else {
			pairs = append(pairs, k+"="+v)
		}
	}

	return pairs
}

// ValidateTextKey if k is not a valid TXT record key.
//
// See https://tools.ietf.org/html/rfc6763#section-6.4
func ValidateTextKey(k string) error {
	// TODO(jmalloc): actually validate
	return nil
}

// ValidateTextValue if v is not a valid TXT record value.
//
//https://tools.ietf.org/html/rfc6763#section-6.5
func ValidateTextValue(v string) error {
	// TODO(jmalloc): actually validate
	panic("ni")
}
