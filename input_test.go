package huhx

import (
	"testing"

	"charm.land/huh/v2"
)

func TestInput_Forwarders(t *testing.T) {
	i := NewInput().
		Key("k").
		Title("t").
		TitleFunc(func() string { return "tf" }, nil).
		Description("d").
		DescriptionFunc(func() string { return "df" }, nil).
		Placeholder("p").
		PlaceholderFunc(func() string { return "pf" }, nil).
		CharLimit(50).
		Prompt("> ").
		Suggestions([]string{"a", "b"}).
		SuggestionsFunc(func() []string { return []string{"c"} }, nil).
		EchoMode(huh.EchoMode(0)).
		Password(true).
		Inline(true)
	if i == nil {
		t.Fatal("expected non-nil input after forwarder chain")
	}
}

func TestInput_AccessorWritesValue(t *testing.T) {
	var dst string
	acc := huh.NewPointerAccessor(&dst)
	form := NewForm(NewGroup(NewInput().Key("name").Accessor(acc)))
	r := New(form, WithNonInteractive(Always), WithAnswers(map[string]any{"name": "alice"}))
	if err := r.Run(); err != nil {
		t.Fatal(err)
	}
	if dst != "alice" {
		t.Errorf("expected accessor to receive value, got %q", dst)
	}
}
