package huhless

import (
	"fmt"

	"charm.land/huh/v2"
)

// Select wraps *huh.Select[T] for headless drive.
type Select[T comparable] struct {
	inner    *huh.Select[T]
	k        string
	value    *T
	validate func(T) error
	optional bool
}

// NewSelect returns a new Select wrapping huh.NewSelect[T]().
func NewSelect[T comparable]() *Select[T] {
	return &Select[T]{inner: huh.NewSelect[T]()}
}

// Key sets the field key used for answer lookup.
func (s *Select[T]) Key(k string) *Select[T] {
	s.k = k
	s.inner.Key(k)
	return s
}

// Title sets the field title.
func (s *Select[T]) Title(t string) *Select[T] {
	s.inner.Title(t)
	return s
}

// Description sets the field description.
func (s *Select[T]) Description(d string) *Select[T] {
	s.inner.Description(d)
	return s
}

// Options sets the available options.
func (s *Select[T]) Options(opts ...huh.Option[T]) *Select[T] {
	s.inner.Options(opts...)
	return s
}

// Value binds a destination pointer.
func (s *Select[T]) Value(v *T) *Select[T] {
	s.value = v
	s.inner.Value(v)
	return s
}

// Validate sets the validator on both the wrapper and the inner huh field.
func (s *Select[T]) Validate(fn func(T) error) *Select[T] {
	s.validate = fn
	s.inner.Validate(fn)
	return s
}

// Optional marks the field as not required in non-interactive mode.
func (s *Select[T]) Optional() *Select[T] {
	s.optional = true
	return s
}

func (s *Select[T]) key() string         { return s.k }
func (s *Select[T]) huhField() huh.Field { return s.inner }
func (s *Select[T]) required() bool      { return !s.optional }

// set resolves the answer string against the field's options. An option
// matches if its Key equals the answer or if fmt.Sprintf("%v", o.Value)
// equals the answer. The %v fallback works cleanly for primitive T
// (string, int, etc.) but may be ambiguous for struct types or types
// with a custom String() method — set Key explicitly for those.
func (s *Select[T]) set(value string) error {
	for _, o := range s.inner.GetOptions() {
		if o.Key == value || fmt.Sprintf("%v", o.Value) == value {
			if s.value != nil {
				*s.value = o.Value
			}
			if s.validate != nil {
				if err := s.validate(o.Value); err != nil {
					return err
				}
			}
			return nil
		}
	}
	return fmt.Errorf("%q is not a valid option", value)
}
