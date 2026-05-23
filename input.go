package huhless

import "charm.land/huh/v2"

// Input is a thin builder over *huh.Input that captures the validator and
// destination pointer so the non-interactive runner can drive the field
// without going through huh's bubble tea loop.
type Input struct {
	inner    *huh.Input
	k        string
	value    *string
	validate func(string) error
	optional bool
}

// NewInput returns a new Input wrapping huh.NewInput().
func NewInput() *Input {
	return &Input{inner: huh.NewInput()}
}

// Key sets the field key used for answer lookup.
func (i *Input) Key(k string) *Input {
	i.k = k
	i.inner.Key(k)
	return i
}

// Title sets the field title.
func (i *Input) Title(s string) *Input {
	i.inner.Title(s)
	return i
}

// Description sets the field description.
func (i *Input) Description(s string) *Input {
	i.inner.Description(s)
	return i
}

// Placeholder sets the field placeholder.
func (i *Input) Placeholder(s string) *Input {
	i.inner.Placeholder(s)
	return i
}

// CharLimit sets the maximum character length.
func (i *Input) CharLimit(n int) *Input {
	i.inner.CharLimit(n)
	return i
}

// Value binds a destination string pointer.
func (i *Input) Value(v *string) *Input {
	i.value = v
	i.inner.Value(v)
	return i
}

// Validate sets the validator on both the wrapper and the inner huh field.
func (i *Input) Validate(fn func(string) error) *Input {
	i.validate = fn
	i.inner.Validate(fn)
	return i
}

// Optional marks the field as not required in non-interactive mode.
func (i *Input) Optional() *Input {
	i.optional = true
	return i
}

func (i *Input) key() string        { return i.k }
func (i *Input) huhField() huh.Field { return i.inner }
func (i *Input) required() bool     { return !i.optional }

func (i *Input) set(value string) error {
	if i.value != nil {
		*i.value = value
	}
	if i.validate != nil {
		if err := i.validate(value); err != nil {
			return err
		}
	}
	return nil
}
