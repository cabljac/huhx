package huhx

import "charm.land/huh/v2"

// Text wraps *huh.Text for headless drive.
type Text struct {
	inner    *huh.Text
	k        string
	value    *string
	validate func(string) error
	optional bool
}

// NewText returns a new Text wrapping huh.NewText().
func NewText() *Text {
	return &Text{inner: huh.NewText()}
}

// Key sets the field key used for answer lookup.
func (t *Text) Key(k string) *Text {
	t.k = k
	t.inner.Key(k)
	return t
}

// Title sets the field title.
func (t *Text) Title(s string) *Text {
	t.inner.Title(s)
	return t
}

// Description sets the field description.
func (t *Text) Description(s string) *Text {
	t.inner.Description(s)
	return t
}

// Placeholder sets the field placeholder.
func (t *Text) Placeholder(s string) *Text {
	t.inner.Placeholder(s)
	return t
}

// CharLimit sets the maximum character length.
func (t *Text) CharLimit(n int) *Text {
	t.inner.CharLimit(n)
	return t
}

// Value binds a destination string pointer.
func (t *Text) Value(v *string) *Text {
	t.value = v
	t.inner.Value(v)
	return t
}

// Validate sets the validator on both the wrapper and the inner huh field.
func (t *Text) Validate(fn func(string) error) *Text {
	t.validate = fn
	t.inner.Validate(fn)
	return t
}

// Optional marks the field as not required in non-interactive mode.
func (t *Text) Optional() *Text {
	t.optional = true
	return t
}

func (t *Text) key() string         { return t.k }
func (t *Text) huhField() huh.Field { return t.inner }
func (t *Text) required() bool      { return !t.optional }

func (t *Text) set(value string) error {
	if t.value != nil {
		*t.value = value
	}
	if t.validate != nil {
		if err := t.validate(value); err != nil {
			return err
		}
	}
	return nil
}
