package huhx

import "charm.land/huh/v2"

// Input is a thin builder over *huh.Input that captures the validator and
// destination pointer so the non-interactive runner can drive the field
// without going through huh's bubble tea loop.
type Input struct {
	inner    *huh.Input
	k        string
	value    *string
	accessor huh.Accessor[string]
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
	i.accessor = huh.NewPointerAccessor(v)
	i.inner.Value(v)
	return i
}

// Accessor sets a custom accessor for reading and writing the field value.
func (i *Input) Accessor(a huh.Accessor[string]) *Input {
	i.accessor = a
	i.inner.Accessor(a)
	return i
}

// TitleFunc sets a dynamic title function.
func (i *Input) TitleFunc(f func() string, bindings any) *Input {
	i.inner.TitleFunc(f, bindings)
	return i
}

// DescriptionFunc sets a dynamic description function.
func (i *Input) DescriptionFunc(f func() string, bindings any) *Input {
	i.inner.DescriptionFunc(f, bindings)
	return i
}

// Prompt sets the input prompt.
func (i *Input) Prompt(s string) *Input {
	i.inner.Prompt(s)
	return i
}

// Suggestions sets the autocomplete suggestions.
func (i *Input) Suggestions(suggestions []string) *Input {
	i.inner.Suggestions(suggestions)
	return i
}

// SuggestionsFunc sets a dynamic suggestions function.
func (i *Input) SuggestionsFunc(f func() []string, bindings any) *Input {
	i.inner.SuggestionsFunc(f, bindings)
	return i
}

// EchoMode sets the echo mode for the input.
func (i *Input) EchoMode(mode huh.EchoMode) *Input {
	i.inner.EchoMode(mode)
	return i
}

// Password toggles password masking for the input.
func (i *Input) Password(password bool) *Input {
	i.inner.Password(password)
	return i
}

// PlaceholderFunc sets a dynamic placeholder function.
func (i *Input) PlaceholderFunc(f func() string, bindings any) *Input {
	i.inner.PlaceholderFunc(f, bindings)
	return i
}

// Inline toggles inline rendering for the input.
func (i *Input) Inline(inline bool) *Input {
	i.inner.Inline(inline)
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
	if i.accessor != nil {
		i.accessor.Set(value)
	} else if i.value != nil {
		*i.value = value
	}
	if i.validate != nil {
		if err := i.validate(value); err != nil {
			return err
		}
	}
	return nil
}
