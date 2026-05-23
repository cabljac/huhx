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
