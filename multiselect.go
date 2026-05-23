package huhx

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
)

// MultiSelect wraps *huh.MultiSelect[T] for headless drive. The non-interactive
// answer is a comma-separated string of option keys or values.
type MultiSelect[T comparable] struct {
	inner    *huh.MultiSelect[T]
	k        string
	value    *[]T
	validate func([]T) error
	optional bool
}

// NewMultiSelect returns a new MultiSelect wrapping huh.NewMultiSelect[T]().
func NewMultiSelect[T comparable]() *MultiSelect[T] {
	return &MultiSelect[T]{inner: huh.NewMultiSelect[T]()}
}

// Key sets the field key used for answer lookup.
func (m *MultiSelect[T]) Key(k string) *MultiSelect[T] {
	m.k = k
	m.inner.Key(k)
	return m
}

// Title sets the field title.
func (m *MultiSelect[T]) Title(t string) *MultiSelect[T] {
	m.inner.Title(t)
	return m
}

// Description sets the field description.
func (m *MultiSelect[T]) Description(d string) *MultiSelect[T] {
	m.inner.Description(d)
	return m
}

// Options sets the available options.
func (m *MultiSelect[T]) Options(opts ...huh.Option[T]) *MultiSelect[T] {
	m.inner.Options(opts...)
	return m
}

// Limit sets the maximum number of selections.
func (m *MultiSelect[T]) Limit(n int) *MultiSelect[T] {
	m.inner.Limit(n)
	return m
}

// Value binds a destination slice pointer.
func (m *MultiSelect[T]) Value(v *[]T) *MultiSelect[T] {
	m.value = v
	m.inner.Value(v)
	return m
}

// Validate sets the validator on both the wrapper and the inner huh field.
func (m *MultiSelect[T]) Validate(fn func([]T) error) *MultiSelect[T] {
	m.validate = fn
	m.inner.Validate(fn)
	return m
}

// Optional marks the field as not required in non-interactive mode.
func (m *MultiSelect[T]) Optional() *MultiSelect[T] {
	m.optional = true
	return m
}

func (m *MultiSelect[T]) key() string         { return m.k }
func (m *MultiSelect[T]) huhField() huh.Field { return m.inner }
func (m *MultiSelect[T]) required() bool      { return !m.optional }

// set parses value as a comma-separated list of option references. Each
// part matches an option whose Key equals it or whose Value formatted as
// fmt.Sprintf("%v", v) equals it. The %v fallback works cleanly for
// primitive T but may be ambiguous for struct types or types with a
// custom String() method — set Key explicitly for those.
func (m *MultiSelect[T]) set(value string) error {
	opts := m.inner.GetOptions()
	parts := strings.Split(value, ",")
	result := make([]T, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		var matched bool
		for _, o := range opts {
			if o.Key == p || fmt.Sprintf("%v", o.Value) == p {
				result = append(result, o.Value)
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("%q is not a valid option", p)
		}
	}
	if m.value != nil {
		*m.value = result
	}
	if m.validate != nil {
		if err := m.validate(result); err != nil {
			return err
		}
	}
	return nil
}
