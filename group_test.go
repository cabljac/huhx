package huhx

import (
	"strings"
	"testing"

	"charm.land/huh/v2"
)

func TestGroup_Forwarders(t *testing.T) {
	g := NewGroup(NewInput().Key("k")).
		Title("t").
		Description("d").
		WithShowHelp(true).
		WithShowErrors(false).
		WithTheme(huh.ThemeFunc(huh.ThemeCharm)).
		WithKeyMap(&huh.KeyMap{}).
		WithWidth(80).
		WithHeight(24)
	if g == nil {
		t.Fatal("expected non-nil group after forwarder chain")
	}
}

func TestGroup_WithHideBool(t *testing.T) {
	build := func(name, secret *string, hidden bool) *Form {
		return NewForm(
			NewGroup(
				NewInput().Key("name").Title("Name").Value(name),
			),
			NewGroup(
				NewInput().Key("secret").Title("Secret").Value(secret),
			).WithHide(hidden),
		)
	}

	t.Run("hidden=true", func(t *testing.T) {
		var name, secret string
		form := build(&name, &secret, true)

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"name": "myapp"}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error when group hidden, got %v", err)
		}
		if name != "myapp" {
			t.Errorf("expected name=%q, got %q", "myapp", name)
		}
		if secret != "" {
			t.Errorf("expected hidden field untouched, got secret=%q", secret)
		}
	})

	t.Run("hidden=false", func(t *testing.T) {
		var name, secret string
		form := build(&name, &secret, false)

		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"name": "myapp"}),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected missing-required error when group visible")
		}
		if !strings.Contains(err.Error(), "missing required answers for:") {
			t.Errorf("expected missing-answers header, got %q", err.Error())
		}
		if !strings.Contains(err.Error(), "--secret") {
			t.Errorf("expected --secret listed, got %q", err.Error())
		}
	})
}
