package huhx

import "charm.land/huh/v2"

// field is the internal interface every huhx field wrapper satisfies.
// It is used by the non-interactive runner to inject answers and surface
// validation errors without touching huh internals.
type field interface {
	key() string
	set(value string) error
	huhField() huh.Field
	required() bool
}
