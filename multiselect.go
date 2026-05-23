package huhx

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
)

// MultiSelect wraps *huh.MultiSelect[T] for headless drive. The
// non-interactive answer is a comma-separated string of option keys or
// values. Options provided via Options(...) are captured at construction
// time; options provided via OptionsFunc(...) are re-evaluated lazily
// inside set() so closures that depend on earlier fields' values resolve
// correctly.
type MultiSelect[T comparable] struct {
	inner       *huh.MultiSelect[T]
	k           string
	value       *[]T
	validate    func([]T) error
	optional    bool
	options     []huh.Option[T]
	optionsFunc func() []huh.Option[T]
	accessor    huh.Accessor[[]T]
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

// TitleFunc sets a dynamic title provider re-evaluated on bindings change.
func (m *MultiSelect[T]) TitleFunc(f func() string, bindings any) *MultiSelect[T] {
	m.inner.TitleFunc(f, bindings)
	return m
}

// DescriptionFunc sets a dynamic description provider re-evaluated on bindings change.
func (m *MultiSelect[T]) DescriptionFunc(f func() string, bindings any) *MultiSelect[T] {
	m.inner.DescriptionFunc(f, bindings)
	return m
}

// Filterable toggles filtering UI.
func (m *MultiSelect[T]) Filterable(filterable bool) *MultiSelect[T] {
	m.inner.Filterable(filterable)
	return m
}

// Filtering sets the current filtering state.
func (m *MultiSelect[T]) Filtering(filtering bool) *MultiSelect[T] {
	m.inner.Filtering(filtering)
	return m
}

// Width sets the field width.
func (m *MultiSelect[T]) Width(w int) *MultiSelect[T] {
	m.inner.Width(w)
	return m
}

// Height sets the field height.
func (m *MultiSelect[T]) Height(h int) *MultiSelect[T] {
	m.inner.Height(h)
	return m
}

// Options sets the available options statically. The wrapper retains a
// copy so the non-interactive runner can match answers without going
// through huh internals. Calling Options clears any previously set
// OptionsFunc.
func (m *MultiSelect[T]) Options(opts ...huh.Option[T]) *MultiSelect[T] {
	m.options = opts
	m.optionsFunc = nil
	m.inner.Options(opts...)
	return m
}

// OptionsFunc sets a dynamic options provider. The provider is forwarded
// to huh for interactive mode and is re-evaluated by the non-interactive
// runner at injection time so closures that depend on earlier fields'
// values resolve correctly. The dependent field must live in a later
// group than its source field (same rule as huh's interactive bindings).
// Calling OptionsFunc clears any previously set static Options.
func (m *MultiSelect[T]) OptionsFunc(f func() []huh.Option[T], bindings any) *MultiSelect[T] {
	m.optionsFunc = f
	m.options = nil
	m.inner.OptionsFunc(f, bindings)
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
	m.accessor = huh.NewPointerAccessor(v)
	m.inner.Value(v)
	return m
}

// Accessor binds a custom accessor for reading and writing the value.
func (m *MultiSelect[T]) Accessor(a huh.Accessor[[]T]) *MultiSelect[T] {
	m.accessor = a
	m.inner.Accessor(a)
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

// currentOptions returns the option slice to match against. If
// OptionsFunc was set the closure is invoked now so it can read any
// earlier-field values that have already been written by the runner.
func (m *MultiSelect[T]) currentOptions() []huh.Option[T] {
	if m.optionsFunc != nil {
		return m.optionsFunc()
	}
	return m.options
}

// set parses value as a comma-separated list of option references. Each
// part matches an option whose Key equals it or whose Value formatted as
// fmt.Sprintf("%v", v) equals it. The %v fallback works cleanly for
// primitive T but may be ambiguous for struct types or types with a
// custom String() method — set Key explicitly for those.
func (m *MultiSelect[T]) set(value string) error {
	opts := m.currentOptions()
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
	if m.accessor != nil {
		m.accessor.Set(result)
	} else if m.value != nil {
		*m.value = result
	}
	if m.validate != nil {
		if err := m.validate(result); err != nil {
			return err
		}
	}
	return nil
}
