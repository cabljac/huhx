package huhx

import (
	"errors"
	"slices"
	"testing"

	"charm.land/huh/v2"
)

type boxAccessor[T any] struct{ v T }

func (b *boxAccessor[T]) Get() T  { return b.v }
func (b *boxAccessor[T]) Set(v T) { b.v = v }

func TestRunner_ValidatorErrorKeepsPriorValue(t *testing.T) {
	errNope := errors.New("nope")
	opts := []huh.Option[string]{huh.NewOption("a", "a"), huh.NewOption("b", "b")}

	type bound struct {
		field field
		kept  func() bool
	}

	cases := []struct {
		name   string
		answer string
		build  func(viaAccessor bool) bound
	}{
		{"input", "bad", func(viaAccessor bool) bound {
			f := NewInput().Key("k").Validate(func(string) error { return errNope })
			if viaAccessor {
				a := &boxAccessor[string]{v: "orig"}
				return bound{f.Accessor(a), func() bool { return a.v == "orig" }}
			}
			v := "orig"
			return bound{f.Value(&v), func() bool { return v == "orig" }}
		}},
		{"text", "bad", func(viaAccessor bool) bound {
			f := NewText().Key("k").Validate(func(string) error { return errNope })
			if viaAccessor {
				a := &boxAccessor[string]{v: "orig"}
				return bound{f.Accessor(a), func() bool { return a.v == "orig" }}
			}
			v := "orig"
			return bound{f.Value(&v), func() bool { return v == "orig" }}
		}},
		{"confirm", "true", func(viaAccessor bool) bound {
			f := NewConfirm().Key("k").Validate(func(bool) error { return errNope })
			if viaAccessor {
				a := &boxAccessor[bool]{}
				return bound{f.Accessor(a), func() bool { return !a.v }}
			}
			v := false
			return bound{f.Value(&v), func() bool { return !v }}
		}},
		{"select", "b", func(viaAccessor bool) bound {
			f := NewSelect[string]().Key("k").Options(opts...).Validate(func(string) error { return errNope })
			if viaAccessor {
				a := &boxAccessor[string]{v: "a"}
				return bound{f.Accessor(a), func() bool { return a.v == "a" }}
			}
			v := "a"
			return bound{f.Value(&v), func() bool { return v == "a" }}
		}},
		{"multiselect", "b", func(viaAccessor bool) bound {
			f := NewMultiSelect[string]().Key("k").Options(opts...).Validate(func([]string) error { return errNope })
			if viaAccessor {
				a := &boxAccessor[[]string]{v: []string{"a"}}
				return bound{f.Accessor(a), func() bool { return slices.Equal(a.v, []string{"a"}) }}
			}
			v := []string{"a"}
			return bound{f.Value(&v), func() bool { return slices.Equal(v, []string{"a"}) }}
		}},
	}

	for _, tc := range cases {
		for _, binding := range []string{"value", "accessor"} {
			t.Run(tc.name+"/"+binding, func(t *testing.T) {
				b := tc.build(binding == "accessor")
				r := New(NewForm(NewGroup(b.field)),
					WithNonInteractive(Always),
					WithAnswers(map[string]any{"k": tc.answer}),
				)
				err := r.Run()
				if !errors.Is(err, errNope) {
					t.Fatalf("expected validator error, got %v", err)
				}
				if !b.kept() {
					t.Errorf("bound value was overwritten by a rejected answer")
				}
			})
		}
	}
}
