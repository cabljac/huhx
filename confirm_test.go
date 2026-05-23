package huhx

import (
	"errors"
	"strings"
	"testing"
)

func TestRunner_ConfirmInvalidBool(t *testing.T) {
	var shipIt bool
	form := NewForm(NewGroup(
		NewConfirm().Key("ship-it").Title("Ship it?").Value(&shipIt),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"ship-it": "maybe"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for unparseable bool")
	}
	if !strings.Contains(err.Error(), `field "ship-it"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), `invalid bool "maybe"`) {
		t.Errorf("expected invalid-bool message, got %q", err.Error())
	}
}

func TestRunner_ConfirmValidatorOnInjected(t *testing.T) {
	var shipIt bool
	want := errors.New("confirm-rejected")

	form := NewForm(NewGroup(
		NewConfirm().Key("ship-it").Value(&shipIt).
			Validate(func(b bool) error {
				if !b {
					return want
				}
				return nil
			}),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"ship-it": "false"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected validator error")
	}
	if !strings.Contains(err.Error(), `field "ship-it"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !errors.Is(err, want) {
		t.Errorf("expected validator sentinel wrapped, got %v", err)
	}
}
