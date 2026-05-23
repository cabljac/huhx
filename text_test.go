package huhx

import (
	"errors"
	"strings"
	"testing"
)

func TestText_E2E(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		var bio string
		form := NewForm(NewGroup(
			NewText().Key("bio").Title("Bio").Value(&bio),
		))

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"bio": "hello world"}),
		)

		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if bio != "hello world" {
			t.Errorf("expected bio=%q, got %q", "hello world", bio)
		}
	})

	t.Run("missing required answer", func(t *testing.T) {
		var bio string
		form := NewForm(NewGroup(
			NewText().Key("bio").Title("Bio").Value(&bio),
		))

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{}),
		)

		err := r.Run()
		if err == nil {
			t.Fatal("expected error for missing required answer")
		}
		msg := err.Error()
		if !strings.Contains(msg, "missing required answers for:") {
			t.Errorf("expected missing-answers header, got %q", msg)
		}
		if !strings.Contains(msg, "--bio") {
			t.Errorf("expected --bio listed, got %q", msg)
		}
	})

	t.Run("validator fires", func(t *testing.T) {
		var bio string
		form := NewForm(NewGroup(
			NewText().Key("bio").Title("Bio").Value(&bio).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("bio is required")
					}
					return nil
				}),
		))

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"bio": ""}),
		)

		err := r.Run()
		if err == nil {
			t.Fatal("expected validator error")
		}
		if !strings.Contains(err.Error(), `field "bio"`) {
			t.Errorf("expected field-prefixed error, got %q", err.Error())
		}
		if !strings.Contains(err.Error(), "bio is required") {
			t.Errorf("expected validator message, got %q", err.Error())
		}
	})

	t.Run("optional omitted", func(t *testing.T) {
		var bio string
		form := NewForm(NewGroup(
			NewText().Key("bio").Title("Bio").Value(&bio).Optional(),
		))

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{}),
		)

		if err := r.Run(); err != nil {
			t.Fatalf("expected no error with optional field omitted, got %v", err)
		}
		if bio != "" {
			t.Errorf("expected bio zero, got %q", bio)
		}
	})
}
