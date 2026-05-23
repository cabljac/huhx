package huhx

import (
	"fmt"
	"strconv"

	"charm.land/huh/v2"
)

// Confirm wraps *huh.Confirm for headless drive.
type Confirm struct {
	inner    *huh.Confirm
	k        string
	value    *bool
	validate func(bool) error
	optional bool
}

// NewConfirm returns a new Confirm wrapping huh.NewConfirm().
func NewConfirm() *Confirm {
	return &Confirm{inner: huh.NewConfirm()}
}

// Key sets the field key used for answer lookup.
func (c *Confirm) Key(k string) *Confirm {
	c.k = k
	c.inner.Key(k)
	return c
}

// Title sets the field title.
func (c *Confirm) Title(s string) *Confirm {
	c.inner.Title(s)
	return c
}

// Description sets the field description.
func (c *Confirm) Description(s string) *Confirm {
	c.inner.Description(s)
	return c
}

// Affirmative sets the affirmative label.
func (c *Confirm) Affirmative(s string) *Confirm {
	c.inner.Affirmative(s)
	return c
}

// Negative sets the negative label.
func (c *Confirm) Negative(s string) *Confirm {
	c.inner.Negative(s)
	return c
}

// Value binds a destination bool pointer.
func (c *Confirm) Value(v *bool) *Confirm {
	c.value = v
	c.inner.Value(v)
	return c
}

// Validate sets the validator on both the wrapper and the inner huh field.
func (c *Confirm) Validate(fn func(bool) error) *Confirm {
	c.validate = fn
	c.inner.Validate(fn)
	return c
}

// Optional marks the field as not required in non-interactive mode.
func (c *Confirm) Optional() *Confirm {
	c.optional = true
	return c
}

func (c *Confirm) key() string         { return c.k }
func (c *Confirm) huhField() huh.Field { return c.inner }
func (c *Confirm) required() bool      { return !c.optional }

func (c *Confirm) set(value string) error {
	b, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("invalid bool %q: %w", value, err)
	}
	if c.value != nil {
		*c.value = b
	}
	if c.validate != nil {
		if err := c.validate(b); err != nil {
			return err
		}
	}
	return nil
}
