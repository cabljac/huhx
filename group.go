package huhless

import "charm.land/huh/v2"

// Group wraps *huh.Group and tracks the hide predicate alongside the
// huhless field list so the non-interactive runner can walk groups in
// order, skip hidden ones, and dispatch answers to the correct field.
type Group struct {
	inner  *huh.Group
	hide   func() bool
	fields []field
}

// NewGroup builds a Group from huhless field wrappers.
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

// WithHide marks the group as skipped (in non-interactive mode) and
// hidden (in interactive mode) when fn returns true. Mirrors huh's
// WithHideFunc naming.
func (g *Group) WithHide(fn func() bool) *Group {
	g.hide = fn
	g.inner.WithHideFunc(fn)
	return g
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
