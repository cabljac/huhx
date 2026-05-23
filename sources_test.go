package huhx

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestRunner_AnswerPairMalformed(t *testing.T) {
	var name string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
	))

	cmd := &cobra.Command{Use: "t"}
	cmd.Flags().StringSlice("answer", nil, "")
	if err := cmd.ParseFlags([]string{"--answer", "name-only-no-equals"}); err != nil {
		t.Fatal(err)
	}

	r := New(form,
		WithNonInteractive(Always),
		WithCobraFlags(cmd),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for malformed --answer pair")
	}
	msg := err.Error()
	if !strings.Contains(msg, `invalid --answer "name-only-no-equals"`) {
		t.Errorf("expected invalid-pair message, got %q", msg)
	}
	if !strings.Contains(msg, "expected key=val") {
		t.Errorf("expected key=val hint, got %q", msg)
	}
}

func TestRunner_AnswerFileNotFound(t *testing.T) {
	var name string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
	))

	bogus := "/nonexistent/path/that/should/not/exist.yaml"
	r := New(form,
		WithNonInteractive(Always),
		WithAnswerFile(bogus),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for missing answer file")
	}
	msg := err.Error()
	if !strings.Contains(msg, "read answer file") {
		t.Errorf("expected read-answer-file prefix, got %q", msg)
	}
	if !strings.Contains(msg, bogus) {
		t.Errorf("expected bogus path in error, got %q", msg)
	}
}
