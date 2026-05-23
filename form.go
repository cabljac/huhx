package huhless

import "charm.land/huh/v2"

// Form wraps *huh.Form and retains the huhless groups so the runner can
// walk them in non-interactive mode.
type Form struct {
	inner  *huh.Form
	groups []*Group
}

// NewForm builds a Form from huhless groups.
func NewForm(groups ...*Group) *Form {
	huhGroups := make([]*huh.Group, len(groups))
	for i, g := range groups {
		huhGroups[i] = g.inner
	}
	return &Form{
		inner:  huh.NewForm(huhGroups...),
		groups: groups,
	}
}

// Huh returns the underlying *huh.Form for callers that need huh-native
// access (e.g. tests, advanced theming).
//
// Calling Run() directly on the returned form bypasses runner mode
// detection and the non-interactive answer pipeline.
func (f *Form) Huh() *huh.Form {
	return f.inner
}
