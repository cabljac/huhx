package huhx

import "charm.land/huh/v2"

// Group wraps *huh.Group and tracks the hide predicate alongside the
// huhx field list so the non-interactive runner can walk groups in
// order, skip hidden ones, and dispatch answers to the correct field.
type Group struct {
	inner  *huh.Group
	hide   func() bool
	fields []field
}

// NewGroup builds a Group from huhx field wrappers.
func NewGroup(fields ...field) *Group {
	huhFields := make([]huh.Field, len(fields))
	for i, f := range fields {
		huhFields[i] = f.huhField()
	}
	return &Group{
		inner:  huh.NewGroup(huhFields...),
		fields: fields,
	}
}

// Title sets the group title.
func (g *Group) Title(s string) *Group {
	g.inner.Title(s)
	return g
}

// Description sets the group description.
func (g *Group) Description(s string) *Group {
	g.inner.Description(s)
	return g
}

// WithShowHelp toggles the help text for the group.
func (g *Group) WithShowHelp(show bool) *Group {
	g.inner.WithShowHelp(show)
	return g
}

// WithShowErrors toggles error display for the group.
func (g *Group) WithShowErrors(show bool) *Group {
	g.inner.WithShowErrors(show)
	return g
}

// WithTheme applies a theme to the group.
func (g *Group) WithTheme(t huh.Theme) *Group {
	g.inner.WithTheme(t)
	return g
}

// WithKeyMap applies a key map to the group.
func (g *Group) WithKeyMap(k *huh.KeyMap) *Group {
	g.inner.WithKeyMap(k)
	return g
}

// WithWidth sets the group width.
func (g *Group) WithWidth(w int) *Group {
	g.inner.WithWidth(w)
	return g
}

// WithHeight sets the group height.
func (g *Group) WithHeight(h int) *Group {
	g.inner.WithHeight(h)
	return g
}

// WithHide marks the group as skipped (in non-interactive mode) and
// hidden (in interactive mode) when hide is true. The predicate stored
// on the wrapper is a constant returning hide.
func (g *Group) WithHide(hide bool) *Group {
	g.hide = func() bool { return hide }
	g.inner.WithHide(hide)
	return g
}

// WithHideFunc marks the group as skipped (in non-interactive mode) and
// hidden (in interactive mode) when fn returns true. Mirrors huh's
// WithHideFunc.
func (g *Group) WithHideFunc(fn func() bool) *Group {
	g.hide = fn
	g.inner.WithHideFunc(fn)
	return g
}
