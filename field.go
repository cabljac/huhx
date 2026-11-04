package huhx

import (
	"fmt"
	"reflect"

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
	// checkDefault reports whether the field's bound value counts as an
	// answer when no source supplies one, and runs the validator on it.
	checkDefault() (bool, error)
	huhField() huh.Field
	required() bool
}

// scalarString renders a single answer value as a string. Strings pass
// through and other scalars fall back to %v. Lists and maps are rejected
// because their %v rendering is never a meaningful single answer.
func scalarString(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	switch reflect.ValueOf(v).Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return "", fmt.Errorf("expected a single value, got %T", v)
	}
	return fmt.Sprintf("%v", v), nil
}

// matchOption resolves one answer value against opts. A non-string value
// of type T matches an option's Value by ==. Anything else is rendered
// with scalarString and matches an option's Key or its %v-formatted
// Value. The %v match works cleanly for primitive T but may be ambiguous
// for struct types or types with a custom String() method; set Key
// explicitly for those.
func matchOption[T comparable](opts []huh.Option[T], v any) (T, error) {
	var zero T
	// Strings skip the == path even when T is string so option keys keep matching.
	if _, isString := v.(string); !isString {
		if t, ok := v.(T); ok {
			for _, o := range opts {
				if o.Value == t {
					return o.Value, nil
				}
			}
			return zero, fmt.Errorf("%v is not a valid option", v)
		}
	}
	s, err := scalarString(v)
	if err != nil {
		return zero, err
	}
	for _, o := range opts {
		if o.Key == s || fmt.Sprintf("%v", o.Value) == s {
			return o.Value, nil
		}
	}
	return zero, fmt.Errorf("%q is not a valid option", s)
}
