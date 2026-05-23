package huhx

import (
	"fmt"

	"charm.land/huh/v2"
)

// Select wraps *huh.Select[T] for headless drive. Options provided via
// Options(...) are captured at construction time; options provided via
// OptionsFunc(...) are re-evaluated lazily inside set() so closures that
// depend on earlier fields' values resolve correctly.
type Select[T comparable] struct {
	inner       *huh.Select[T]
	k           string
	value       *T
	validate    func(T) error
	optional    bool
	options     []huh.Option[T]
	optionsFunc func() []huh.Option[T]
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

// Options sets the available options statically. The wrapper retains a
// copy so the non-interactive runner can match answers without going
// through huh internals. Calling Options clears any previously set
// OptionsFunc.
func (s *Select[T]) Options(opts ...huh.Option[T]) *Select[T] {
	s.options = opts
	s.optionsFunc = nil
	s.inner.Options(opts...)
	return s
}

// OptionsFunc sets a dynamic options provider. The provider is forwarded
// to huh for interactive mode and is re-evaluated by the non-interactive
// runner at injection time so closures that depend on earlier fields'
// values resolve correctly. The dependent field must live in a later
// group than its source field (same rule as huh's interactive bindings).
// Calling OptionsFunc clears any previously set static Options.
func (s *Select[T]) OptionsFunc(f func() []huh.Option[T], bindings any) *Select[T] {
	s.optionsFunc = f
	s.options = nil
	s.inner.OptionsFunc(f, bindings)
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

// currentOptions returns the option slice to match against. If
// OptionsFunc was set the closure is invoked now so it can read any
// earlier-field values that have already been written by the runner.
func (s *Select[T]) currentOptions() []huh.Option[T] {
	if s.optionsFunc != nil {
		return s.optionsFunc()
	}
	return s.options
}

// set resolves the answer string against the field's options. An option
// matches if its Key equals the answer or if fmt.Sprintf("%v", o.Value)
// equals the answer. The %v fallback works cleanly for primitive T
// (string, int, etc.) but may be ambiguous for struct types or types
// with a custom String() method — set Key explicitly for those.
func (s *Select[T]) set(value string) error {
	for _, o := range s.currentOptions() {
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
