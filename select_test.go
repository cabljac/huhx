package huhx

import (
	"errors"
	"strings"
	"testing"

	"charm.land/huh/v2"
)

// TestRunner_SelectValidatorOnInjected confirms that when an injected
// answer matches an option, the configured validator still runs on the
// resolved Value and a returned error propagates through the runner's
// field-prefixed wrap.
func TestRunner_SelectValidatorOnInjected(t *testing.T) {
	var env string
	want := errors.New("env-rejected")

	form := NewForm(NewGroup(
		NewSelect[string]().Key("env").
			Options(
				huh.NewOption("staging", "staging"),
				huh.NewOption("prod", "prod"),
			).
			Value(&env).
			Validate(func(s string) error {
				if s == "prod" {
					return want
				}
				return nil
			}),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"env": "prod"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected validator error")
	}
	if !strings.Contains(err.Error(), `field "env"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !errors.Is(err, want) {
		t.Errorf("expected validator sentinel wrapped, got %v", err)
	}
}

func TestSelect_Forwarders(t *testing.T) {
	s := NewSelect[string]().
		Key("k").
		Title("t").
		TitleFunc(func() string { return "tf" }, nil).
		Description("d").
		DescriptionFunc(func() string { return "df" }, nil).
		Options(huh.NewOption("a", "a")).
		Filtering(false).
		Inline(true).
		Height(5)
	if s == nil {
		t.Fatal("expected non-nil select after forwarder chain")
	}
}

func TestSelect_AccessorWritesValue(t *testing.T) {
	var dst string
	acc := huh.NewPointerAccessor(&dst)
	form := NewForm(NewGroup(
		NewSelect[string]().
			Key("env").
			Options(huh.NewOption("staging", "staging"), huh.NewOption("prod", "prod")).
			Accessor(acc),
	))
	r := New(form, WithNonInteractive(Always), WithAnswers(map[string]any{"env": "prod"}))
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
	if dst != "prod" {
		t.Errorf("expected accessor to receive %q, got %q", "prod", dst)
	}
}

func TestRunner_SelectTypedAnswer(t *testing.T) {
	t.Run("value of type T matches by ==", func(t *testing.T) {
		var got int
		form := NewForm(NewGroup(
			NewSelect[int]().Key("n").
				Options(huh.NewOption("one", 1), huh.NewOption("two", 2)).
				Value(&got),
		))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"n": 2}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != 2 {
			t.Errorf("expected 2, got %d", got)
		}
	})

	t.Run("value of type T not in options", func(t *testing.T) {
		var got int
		form := NewForm(NewGroup(
			NewSelect[int]().Key("n").
				Options(huh.NewOption("one", 1)).
				Value(&got),
		))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"n": 7}),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected error")
		}
		if !strings.Contains(err.Error(), `field "n": 7 is not a valid option`) {
			t.Errorf("expected invalid-option error, got %q", err.Error())
		}
	})

	t.Run("string still matches by key", func(t *testing.T) {
		var got int
		form := NewForm(NewGroup(
			NewSelect[int]().Key("n").
				Options(huh.NewOption("one", 1), huh.NewOption("two", 2)).
				Value(&got),
		))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"n": "two"}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != 2 {
			t.Errorf("expected 2, got %d", got)
		}
	})

	t.Run("scalar of other type falls back to %v", func(t *testing.T) {
		var got string
		form := NewForm(NewGroup(
			NewSelect[string]().Key("size").
				Options(huh.NewOption("small", "1"), huh.NewOption("large", "3")).
				Value(&got),
		))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"size": 3}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "3" {
			t.Errorf("expected %q, got %q", "3", got)
		}
	})

	t.Run("list is rejected", func(t *testing.T) {
		var got string
		form := NewForm(NewGroup(
			NewSelect[string]().Key("env").
				Options(huh.NewOption("prod", "prod")).
				Value(&got),
		))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"env": []string{"prod"}}),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected error for list answer")
		}
		if !strings.Contains(err.Error(), `field "env": expected a single value, got []string`) {
			t.Errorf("expected single-value error, got %q", err.Error())
		}
	})
}
