package huhx

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"charm.land/huh/v2"
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

func writeAnswerFile(t *testing.T, name, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunner_AnswerFileTypedValues(t *testing.T) {
	cases := map[string]struct {
		file string
		body string
	}{
		"YAML flow list":  {"a.yaml", "regions: [us-east-1, us-west-2]\nport: 8080\nsize: 3\n"},
		"YAML block list": {"a.yaml", "regions:\n  - us-east-1\n  - us-west-2\nport: 8080\nsize: 3\n"},
		"YAML comma":      {"a.yaml", "regions: us-east-1,us-west-2\nport: \"8080\"\nsize: \"3\"\n"},
		"JSON array":      {"a.json", `{"regions": ["us-east-1", "us-west-2"], "port": 8080, "size": 3}`},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var regions []string
			var port, size string
			form := NewForm(NewGroup(
				NewMultiSelect[string]().Key("regions").
					Options(
						huh.NewOption("us-east-1", "us-east-1"),
						huh.NewOption("us-west-2", "us-west-2"),
					).
					Value(&regions),
				NewInput().Key("port").Value(&port),
				NewSelect[string]().Key("size").
					Options(huh.NewOption("small", "1"), huh.NewOption("large", "3")).
					Value(&size),
			))

			r := New(form,
				WithNonInteractive(Always),
				WithAnswerFile(writeAnswerFile(t, tc.file, tc.body)),
			)
			if err := r.Run(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if want := []string{"us-east-1", "us-west-2"}; !reflect.DeepEqual(regions, want) {
				t.Errorf("expected regions %v, got %v", want, regions)
			}
			if port != "8080" {
				t.Errorf("expected port %q, got %q", "8080", port)
			}
			if size != "3" {
				t.Errorf("expected size %q, got %q", "3", size)
			}
		})
	}
}

func TestRunner_AnswerFileNestedMap(t *testing.T) {
	t.Run("rejected when the key is used", func(t *testing.T) {
		var name string
		form := NewForm(NewGroup(NewInput().Key("name").Value(&name)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswerFile(writeAnswerFile(t, "a.yaml", "name:\n  first: a\n")),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected error for nested map answer")
		}
		if !strings.Contains(err.Error(), `field "name": expected a single value, got map[string]interface {}`) {
			t.Errorf("expected single-value error naming the key, got %q", err.Error())
		}
		if name != "" {
			t.Errorf("expected bound value untouched, got %q", name)
		}
	})

	t.Run("ignored when no field uses the key", func(t *testing.T) {
		var name string
		form := NewForm(NewGroup(NewInput().Key("name").Value(&name)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswerFile(writeAnswerFile(t, "a.yaml", "name: app\nextra:\n  k: v\n")),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if name != "app" {
			t.Errorf("expected name %q, got %q", "app", name)
		}
	})
}

func TestRunner_AnswerFileNullIsMissing(t *testing.T) {
	t.Run("required field is reported missing", func(t *testing.T) {
		var name string
		form := NewForm(NewGroup(NewInput().Key("name").Value(&name)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswerFile(writeAnswerFile(t, "a.yaml", "name:\n")),
		)
		err := r.Run()
		if err == nil {
			t.Fatal("expected missing-answer error for null value")
		}
		if !strings.Contains(err.Error(), "missing required answers for:") || !strings.Contains(err.Error(), "--name") {
			t.Errorf("expected missing-answer error for name, got %q", err.Error())
		}
		if name != "" {
			t.Errorf("expected name untouched, got %q", name)
		}
	})

	t.Run("env layer below still applies", func(t *testing.T) {
		t.Setenv("APP_NAME", "from-env")
		var name string
		form := NewForm(NewGroup(NewInput().Key("name").Value(&name)))
		r := New(form,
			WithNonInteractive(Always),
			WithEnvPrefix("APP"),
			WithAnswerFile(writeAnswerFile(t, "a.yaml", "name: ~\n")),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if name != "from-env" {
			t.Errorf("expected name %q, got %q", "from-env", name)
		}
	})
}
