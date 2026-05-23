package huhx

import (
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
