package huhx

import (
	"os"
	"path/filepath"
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
	cmd.Flags().StringArray("answer", nil, "")
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

// TestRunner_AnswerFileMalformedYAML exercises loadAnswerFile's parse
// failure path. The file exists and is readable but contains a YAML
// document that yaml.v3 cannot decode into a map.
func TestRunner_AnswerFileMalformedYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	// "name: : :" is well-formed text but produces a mapping value that
	// is itself a mapping with a nil key, which yaml.v3 rejects.
	body := "name: : :\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	var name string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswerFile(path),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for malformed YAML")
	}
	msg := err.Error()
	if !strings.Contains(msg, "parse answer file") {
		t.Errorf("expected parse-answer-file prefix, got %q", msg)
	}
	if !strings.Contains(msg, path) {
		t.Errorf("expected file path in error, got %q", msg)
	}
}

// TestLoadAnswerFile_CacheAndInvalidate verifies the answer file is memoized
// across calls, that each call returns an independent copy, and that the
// cache is invalidated when the file changes.
func TestLoadAnswerFile_CacheAndInvalidate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "answers.yaml")
	if err := os.WriteFile(path, []byte("name: alice\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	first, err := loadAnswerFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if first["name"] != "alice" {
		t.Fatalf("got %q, want alice", first["name"])
	}

	// Mutating the returned copy must not poison the cache.
	first["name"] = "mutated"
	second, err := loadAnswerFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if second["name"] != "alice" {
		t.Fatalf("cache returned a shared/mutated map: got %q, want alice", second["name"])
	}

	// Rewriting the file (different size) invalidates the cache.
	if err := os.WriteFile(path, []byte("name: bob-the-builder\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	third, err := loadAnswerFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if third["name"] != "bob-the-builder" {
		t.Fatalf("cache not invalidated after rewrite: got %q", third["name"])
	}
}
