package huhx

import (
	"fmt"

	"charm.land/huh/v2"
)

// field is the internal interface every huhx field wrapper satisfies.
// It is used by the non-interactive runner to inject answers and surface
// validation errors without touching huh internals. set receives the raw
// answer from whichever source resolved it and coerces it to the field's
// value type.
type field interface {
	key() string
	set(value any) error
	huhField() huh.Field
	required() bool
}

// answerString renders an answer value for string-based parsing.
func answerString(v any) string {
	return fmt.Sprintf("%v", v)
}
