package huhx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"
)

// buildForm constructs the standard "deploy" test form. The hidden flag
// controls whether the second group is skipped via When().
func buildForm(name, environment *string, allRegions *bool, hidden bool) *Form {
	return NewForm(
		NewGroup(
			NewInput().Key("name").Title("App name").Value(name).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("name is required")
					}
					return nil
				}),
			NewSelect[string]().Key("environment").Title("Environment").
				Options(
					huh.NewOption("staging", "staging"),
					huh.NewOption("prod", "prod"),
				).Value(environment),
		),
		NewGroup(
			NewConfirm().Key("all-regions").Title("All regions?").Value(allRegions),
		).WithHideFunc(func() bool { return hidden }),
	)
}

func TestRunner_AllAnswersSupplied(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, false)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"name":        "myapp",
			"environment": "prod",
			"all-regions": "true",
		}),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "myapp" {
		t.Errorf("expected name=%q, got %q", "myapp", name)
	}
	if environment != "prod" {
		t.Errorf("expected environment=%q, got %q", "prod", environment)
	}
	if !allRegions {
		t.Errorf("expected all-regions=true, got %v", allRegions)
	}
}

func TestRunner_MissingRequiredAnswer(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, false)

	r := New(form,
		WithNonInteractive(Always),
		WithEnvPrefix("DEPLOY"),
		WithAnswers(map[string]any{"name": "myapp"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for missing answers")
	}
	msg := err.Error()
	if !strings.Contains(msg, "missing required answers for:") {
		t.Errorf("expected missing-answers header, got %q", msg)
	}
	if !strings.Contains(msg, "--environment") {
		t.Errorf("expected --environment listed, got %q", msg)
	}
	if !strings.Contains(msg, "(env: DEPLOY_ENVIRONMENT)") {
		t.Errorf("expected env hint for environment, got %q", msg)
	}
	if strings.Contains(msg, "--all-regions") {
		t.Errorf("expected Confirm never listed as missing, got %q", msg)
	}
}

func TestRunner_ValidatorFiresOnInjectedValue(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, false)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"name":        "",
			"environment": "prod",
			"all-regions": "false",
		}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected validator error")
	}
	if !strings.Contains(err.Error(), `field "name"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "name is required") {
		t.Errorf("expected validator message, got %q", err.Error())
	}
}

func TestRunner_HiddenGroupSkipped(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, true)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"name":        "myapp",
			"environment": "staging",
		}),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error when hidden group's field is absent, got %v", err)
	}
	if allRegions {
		t.Errorf("expected hidden field untouched, got allRegions=%v", allRegions)
	}
}

func TestRunner_SelectRejectsUnknownOption(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, false)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"name":        "myapp",
			"environment": "production",
			"all-regions": "true",
		}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for invalid select option")
	}
	if !strings.Contains(err.Error(), `field "environment"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), `"production" is not a valid option`) {
		t.Errorf("expected invalid-option message, got %q", err.Error())
	}
}

func TestRunner_PrecedenceAnswerOverridesEnv(t *testing.T) {
	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, true)

	t.Setenv("DEPLOY_NAME", "from-env")
	t.Setenv("DEPLOY_ENVIRONMENT", "staging")

	r := New(form,
		WithNonInteractive(Always),
		WithEnvPrefix("DEPLOY"),
		WithAnswers(map[string]any{"name": "from-answers"}),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "from-answers" {
		t.Errorf("expected name to come from WithAnswers, got %q", name)
	}
	if environment != "staging" {
		t.Errorf("expected environment to fall through to env, got %q", environment)
	}
}

func TestRunner_AnswerFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "answers.yaml")
	body := "name: yaml-name\nenvironment: prod\nall-regions: true\n"
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}

	var name, environment string
	var allRegions bool
	form := buildForm(&name, &environment, &allRegions, false)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswerFile(path),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if name != "yaml-name" || environment != "prod" || !allRegions {
		t.Errorf("unexpected values: name=%q env=%q all=%v", name, environment, allRegions)
	}
}

func TestRunner_FullPrecedenceChain(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "answers.yaml")
	// File omits env-key and default-key so each downstream source has
	// exactly one layer above it providing a value.
	fileBody := "" +
		"flag-key: from-file\n" +
		"answer-key: from-file\n" +
		"file-key: from-file\n"
	if err := os.WriteFile(path, []byte(fileBody), 0o600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("CHAIN_FLAG_KEY", "from-env")
	t.Setenv("CHAIN_ANSWER_KEY", "from-env")
	t.Setenv("CHAIN_FILE_KEY", "from-env")
	t.Setenv("CHAIN_ENV_KEY", "from-env")
	// Intentionally do not set CHAIN_DEFAULT_KEY so the pointer's existing
	// value survives.

	var flagVal, answerVal, fileVal, envVal, defaultVal string
	defaultVal = "from-default"

	form := NewForm(NewGroup(
		NewInput().Key("flag-key").Value(&flagVal),
		NewInput().Key("answer-key").Value(&answerVal),
		NewInput().Key("file-key").Value(&fileVal),
		NewInput().Key("env-key").Value(&envVal),
		NewInput().Key("default-key").Value(&defaultVal).Optional(),
	))

	cmd := &cobra.Command{Use: "t"}
	cmd.Flags().String("flag-key", "", "")
	cmd.Flags().StringArray("answer", nil, "")
	if err := cmd.ParseFlags([]string{
		"--flag-key", "from-flag",
		"--answer", "answer-key=from-answer",
	}); err != nil {
		t.Fatal(err)
	}

	r := New(form,
		WithNonInteractive(Always),
		WithEnvPrefix("CHAIN"),
		WithCobraFlags(cmd),
		WithAnswerFile(path),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if flagVal != "from-flag" {
		t.Errorf("flag-key: expected named cobra flag to win, got %q", flagVal)
	}
	if answerVal != "from-answer" {
		t.Errorf("answer-key: expected --answer to win over file/env, got %q", answerVal)
	}
	if fileVal != "from-file" {
		t.Errorf("file-key: expected file to win over env, got %q", fileVal)
	}
	if envVal != "from-env" {
		t.Errorf("env-key: expected env fallback, got %q", envVal)
	}
	if defaultVal != "from-default" {
		t.Errorf("default-key: expected pointer default preserved, got %q", defaultVal)
	}
}

func TestRunner_MultiSelect(t *testing.T) {
	var toppings []string
	form := NewForm(NewGroup(
		NewMultiSelect[string]().Key("toppings").
			Options(
				huh.NewOption("cheese", "cheese"),
				huh.NewOption("tomato", "tomato"),
				huh.NewOption("onion", "onion"),
			).Value(&toppings),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"toppings": "cheese, tomato"}),
	)
	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(toppings) != 2 || toppings[0] != "cheese" || toppings[1] != "tomato" {
		t.Errorf("expected [cheese tomato], got %v", toppings)
	}

	toppings = nil
	r = New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"toppings": "cheese,pineapple"}),
	)
	err := r.Run()
	if err == nil {
		t.Fatal("expected error for unknown multiselect option")
	}
	if !strings.Contains(err.Error(), `field "toppings"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), `"pineapple" is not a valid option`) {
		t.Errorf("expected invalid-option message, got %q", err.Error())
	}
}

func TestRunner_Optional(t *testing.T) {
	var name, nickname string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
		NewInput().Key("nickname").Value(&nickname).Optional(),
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"name": "alice"}),
	)
	if err := r.Run(); err != nil {
		t.Fatalf("expected no error with optional field omitted, got %v", err)
	}
	if name != "alice" {
		t.Errorf("expected name=%q, got %q", "alice", name)
	}
	if nickname != "" {
		t.Errorf("expected nickname zero, got %q", nickname)
	}
}

func TestRunner_OptionsFuncDependsOnEarlierField(t *testing.T) {
	var country, state string

	states := map[string][]huh.Option[string]{
		"US": {huh.NewOption("California", "CA"), huh.NewOption("Texas", "TX")},
		"CA": {huh.NewOption("Ontario", "ON"), huh.NewOption("Quebec", "QC")},
	}

	form := NewForm(
		NewGroup(
			NewSelect[string]().Key("country").
				Options(
					huh.NewOption("United States", "US"),
					huh.NewOption("Canada", "CA"),
				).Value(&country),
		),
		NewGroup(
			NewSelect[string]().Key("state").
				OptionsFunc(func() []huh.Option[string] {
					return states[country]
				}, &country).
				Value(&state),
		),
	)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"country": "US",
			"state":   "TX",
		}),
	)
	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if country != "US" {
		t.Errorf("expected country=US, got %q", country)
	}
	if state != "TX" {
		t.Errorf("expected state=TX, got %q", state)
	}
}

func TestRunner_OptionsFuncRejectsUnknown(t *testing.T) {
	var country, state string

	form := NewForm(
		NewGroup(
			NewSelect[string]().Key("country").
				Options(huh.NewOption("United States", "US")).
				Value(&country),
		),
		NewGroup(
			NewSelect[string]().Key("state").
				OptionsFunc(func() []huh.Option[string] {
					if country == "US" {
						return []huh.Option[string]{huh.NewOption("California", "CA")}
					}
					return nil
				}, &country).
				Value(&state),
		),
	)

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{
			"country": "US",
			"state":   "ON",
		}),
	)
	err := r.Run()
	if err == nil {
		t.Fatal("expected error for option not present in dynamic list")
	}
	if !strings.Contains(err.Error(), `field "state"`) {
		t.Errorf("expected field-prefixed error, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), `"ON" is not a valid option`) {
		t.Errorf("expected invalid-option message, got %q", err.Error())
	}
}

func TestRunner_KeylessRequiredFieldErrors(t *testing.T) {
	var name string
	// Second field has no Key(); required by default. Runner must surface
	// a clear error pointing at its position rather than silently treating
	// it as a missing answer with an empty flag name.
	form := NewForm(NewGroup(
		NewInput().Key("first").Value(&name),
		NewInput().Value(&name), // keyless
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"first": "alice"}),
	)

	err := r.Run()
	if err == nil {
		t.Fatal("expected error for keyless required field")
	}
	msg := err.Error()
	if !strings.Contains(msg, "no Key() set") {
		t.Errorf("expected keyless message, got %q", msg)
	}
	if !strings.Contains(msg, "field 2") || !strings.Contains(msg, "group 1") {
		t.Errorf("expected 1-based field location, got %q", msg)
	}
}

func TestRunner_KeylessOptionalFieldSkipped(t *testing.T) {
	var name, nickname string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
		NewInput().Value(&nickname).Optional(), // keyless but optional
	))

	r := New(form,
		WithNonInteractive(Always),
		WithAnswers(map[string]any{"name": "alice"}),
	)

	if err := r.Run(); err != nil {
		t.Fatalf("expected no error for keyless optional field, got %v", err)
	}
	if name != "alice" {
		t.Errorf("expected name=alice, got %q", name)
	}
	if nickname != "" {
		t.Errorf("expected nickname zero, got %q", nickname)
	}
}

func TestRunner_ScalarAnswerFallback(t *testing.T) {
	t.Run("int for Input renders via %v", func(t *testing.T) {
		var got string
		form := NewForm(NewGroup(NewInput().Key("port").Value(&got)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"port": 8080}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "8080" {
			t.Errorf("expected %q, got %q", "8080", got)
		}
	})

	t.Run("bool for Text renders via %v", func(t *testing.T) {
		var got string
		form := NewForm(NewGroup(NewText().Key("notes").Value(&got)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"notes": true}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if got != "true" {
			t.Errorf("expected %q, got %q", "true", got)
		}
	})

	t.Run("bool for Confirm passes through", func(t *testing.T) {
		var got bool
		form := NewForm(NewGroup(NewConfirm().Key("ok").Value(&got)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"ok": true}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !got {
			t.Error("expected true")
		}
	})

	t.Run("int for Confirm parses via %v", func(t *testing.T) {
		var got bool
		form := NewForm(NewGroup(NewConfirm().Key("ok").Value(&got)))
		r := New(form,
			WithNonInteractive(Always),
			WithAnswers(map[string]any{"ok": 1}),
		)
		if err := r.Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !got {
			t.Error("expected true")
		}
	})
}

func TestRunner_ListAnswerRejectedForScalarFields(t *testing.T) {
	cases := map[string]field{
		"Input":   NewInput().Key("x").Value(new(string)),
		"Text":    NewText().Key("x").Value(new(string)),
		"Confirm": NewConfirm().Key("x").Value(new(bool)),
	}
	for name, f := range cases {
		t.Run(name, func(t *testing.T) {
			form := NewForm(NewGroup(f))
			r := New(form,
				WithNonInteractive(Always),
				WithAnswers(map[string]any{"x": []string{"a", "b"}}),
			)
			err := r.Run()
			if err == nil {
				t.Fatal("expected error for list answer")
			}
			if !strings.Contains(err.Error(), `field "x": expected a single value, got []string`) {
				t.Errorf("expected single-value error, got %q", err.Error())
			}
		})
	}
}

func TestRunner_CobraSliceFlagPassesList(t *testing.T) {
	cases := map[string][]string{
		"repeated":        {"--regions", "us-east-1", "--regions", "us-west-2"},
		"comma-separated": {"--regions", "us-east-1,us-west-2"},
	}
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			var regions []string
			form := NewForm(NewGroup(
				NewMultiSelect[string]().Key("regions").
					Options(
						huh.NewOption("us-east-1", "us-east-1"),
						huh.NewOption("us-west-2", "us-west-2"),
					).
					Value(&regions),
			))

			cmd := &cobra.Command{Use: "t"}
			cmd.Flags().StringSlice("regions", nil, "")
			if err := cmd.ParseFlags(args); err != nil {
				t.Fatal(err)
			}

			r := New(form,
				WithNonInteractive(Always),
				WithCobraFlags(cmd),
			)
			if err := r.Run(); err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			want := []string{"us-east-1", "us-west-2"}
			if len(regions) != 2 || regions[0] != want[0] || regions[1] != want[1] {
				t.Errorf("expected %v, got %v", want, regions)
			}
		})
	}
}

func TestRunner_CobraScalarFlagStaysString(t *testing.T) {
	var regions []string
	form := NewForm(NewGroup(
		NewMultiSelect[string]().Key("regions").
			Options(
				huh.NewOption("us-east-1", "us-east-1"),
				huh.NewOption("us-west-2", "us-west-2"),
			).
			Value(&regions),
	))

	cmd := &cobra.Command{Use: "t"}
	cmd.Flags().String("regions", "", "")
	if err := cmd.ParseFlags([]string{"--regions", "us-east-1,us-west-2"}); err != nil {
		t.Fatal(err)
	}

	r := New(form,
		WithNonInteractive(Always),
		WithCobraFlags(cmd),
	)
	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := []string{"us-east-1", "us-west-2"}
	if len(regions) != 2 || regions[0] != want[0] || regions[1] != want[1] {
		t.Errorf("expected %v, got %v", want, regions)
	}
}

func TestRunner_BoundValueIsDefault(t *testing.T) {
	opts := []huh.Option[string]{huh.NewOption("a", "a"), huh.NewOption("b", "b")}

	t.Run("input", func(t *testing.T) {
		name := "prefilled"
		form := NewForm(NewGroup(NewInput().Key("name").Value(&name)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if name != "prefilled" {
			t.Errorf("expected name kept, got %q", name)
		}
	})

	t.Run("text", func(t *testing.T) {
		bio := "prefilled"
		form := NewForm(NewGroup(NewText().Key("bio").Value(&bio)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if bio != "prefilled" {
			t.Errorf("expected bio kept, got %q", bio)
		}
	})

	t.Run("select", func(t *testing.T) {
		env := "b"
		form := NewForm(NewGroup(NewSelect[string]().Key("env").Options(opts...).Value(&env)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env != "b" {
			t.Errorf("expected env kept, got %q", env)
		}
	})

	t.Run("select set after build", func(t *testing.T) {
		var env string
		form := NewForm(NewGroup(NewSelect[string]().Key("env").Options(opts...).Value(&env)))
		env = "b"
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if env != "b" {
			t.Errorf("expected env kept, got %q", env)
		}
	})

	t.Run("select with OptionsFunc", func(t *testing.T) {
		env := "b"
		form := NewForm(NewGroup(NewSelect[string]().Key("env").
			OptionsFunc(func() []huh.Option[string] { return opts }, nil).Value(&env)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("multiselect", func(t *testing.T) {
		regions := []string{"a"}
		form := NewForm(NewGroup(NewMultiSelect[string]().Key("regions").Options(opts...).Value(&regions)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(regions) != 1 || regions[0] != "a" {
			t.Errorf("expected regions kept, got %v", regions)
		}
	})

	t.Run("confirm false", func(t *testing.T) {
		ok := false
		form := NewForm(NewGroup(NewConfirm().Key("ok").Value(&ok)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if ok {
			t.Errorf("expected ok false, got true")
		}
	})

	t.Run("confirm true", func(t *testing.T) {
		ok := true
		form := NewForm(NewGroup(NewConfirm().Key("ok").Value(&ok)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !ok {
			t.Errorf("expected ok true, got false")
		}
	})

	t.Run("confirm unbound", func(t *testing.T) {
		form := NewForm(NewGroup(NewConfirm().Key("ok")))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("read through accessor", func(t *testing.T) {
		a := &boxAccessor[string]{v: "from-accessor"}
		form := NewForm(NewGroup(NewInput().Key("name").Accessor(a)))
		if err := New(form, WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if a.v != "from-accessor" {
			t.Errorf("expected accessor value kept, got %q", a.v)
		}
	})
}

func TestRunner_EmptyBoundValueIsMissing(t *testing.T) {
	opts := []huh.Option[string]{huh.NewOption("a", "a"), huh.NewOption("b", "b")}
	var (
		name, bio string
		env       = "not-an-option"
		dyn       = "a"
		late      string
		regions   []string
	)
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name),
		NewText().Key("bio").Value(&bio),
		NewSelect[string]().Key("env").Options(opts...).Value(&env),
		NewSelect[string]().Key("dyn").
			OptionsFunc(func() []huh.Option[string] { return opts[1:] }, nil).Value(&dyn),
		NewSelect[string]().Key("late").Value(&late).Options(opts...),
		NewMultiSelect[string]().Key("regions").Options(opts...).Value(&regions),
		NewInput().Key("unbound"),
	))
	err := New(form, WithNonInteractive(Always)).Run()
	if err == nil {
		t.Fatal("expected missing-answer error")
	}
	for _, k := range []string{"--name", "--bio", "--env", "--dyn", "--late", "--regions", "--unbound"} {
		if !strings.Contains(err.Error(), k) {
			t.Errorf("expected %s listed, got %q", k, err.Error())
		}
	}
}

func TestRunner_SourcesWinOverBoundValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "answers.yaml")
	if err := os.WriteFile(path, []byte("file-key: from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SRC_ENV_KEY", "from-env")

	vals := map[string]*string{}
	var fields []field
	for _, k := range []string{"map-key", "flag-key", "answer-key", "file-key", "env-key"} {
		v := "from-default"
		vals[k] = &v
		fields = append(fields, NewInput().Key(k).Value(&v))
	}
	ok := true
	fields = append(fields, NewConfirm().Key("confirm-key").Value(&ok))

	cmd := &cobra.Command{Use: "t"}
	cmd.Flags().String("flag-key", "", "")
	cmd.Flags().StringArray("answer", nil, "")
	if err := cmd.ParseFlags([]string{
		"--flag-key", "from-flag",
		"--answer", "answer-key=from-answer",
	}); err != nil {
		t.Fatal(err)
	}

	r := New(NewForm(NewGroup(fields...)),
		WithNonInteractive(Always),
		WithEnvPrefix("SRC"),
		WithCobraFlags(cmd),
		WithAnswerFile(path),
		WithAnswers(map[string]any{"map-key": "from-map", "confirm-key": false}),
	)
	if err := r.Run(); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	want := map[string]string{
		"map-key":    "from-map",
		"flag-key":   "from-flag",
		"answer-key": "from-answer",
		"file-key":   "from-file",
		"env-key":    "from-env",
	}
	for k, w := range want {
		if *vals[k] != w {
			t.Errorf("%s: expected %q, got %q", k, w, *vals[k])
		}
	}
	if ok {
		t.Errorf("confirm-key: expected source false to win over bound true")
	}
}

func TestRunner_ValidatorRunsOnBoundValue(t *testing.T) {
	errBad := errors.New("bad default")
	opts := []huh.Option[string]{huh.NewOption("a", "a"), huh.NewOption("b", "b")}
	name := "prefilled"
	bio := "prefilled"
	env := "a"
	regions := []string{"a"}
	ok := false

	cases := []struct {
		name  string
		field field
	}{
		{"input", NewInput().Key("k").Value(&name).Validate(func(string) error { return errBad })},
		{"text", NewText().Key("k").Value(&bio).Validate(func(string) error { return errBad })},
		{"select", NewSelect[string]().Key("k").Options(opts...).Value(&env).Validate(func(string) error { return errBad })},
		{"multiselect", NewMultiSelect[string]().Key("k").Options(opts...).Value(&regions).Validate(func([]string) error { return errBad })},
		{"confirm", NewConfirm().Key("k").Value(&ok).Validate(func(bool) error { return errBad })},
		{"optional input", NewInput().Key("k").Value(&name).Validate(func(string) error { return errBad }).Optional()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := New(NewForm(NewGroup(tc.field)), WithNonInteractive(Always)).Run()
			if !errors.Is(err, errBad) {
				t.Fatalf("expected validator error, got %v", err)
			}
			if !strings.Contains(err.Error(), `field "k"`) {
				t.Errorf("expected field-prefixed error, got %q", err.Error())
			}
		})
	}

	t.Run("empty optional input skips validator", func(t *testing.T) {
		var empty string
		f := NewInput().Key("k").Value(&empty).Validate(func(string) error { return errBad }).Optional()
		if err := New(NewForm(NewGroup(f)), WithNonInteractive(Always)).Run(); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

type portError struct{ port string }

func (e *portError) Error() string { return "bad port " + e.port }

func TestRunner_ReportsEveryErrorInGroup(t *testing.T) {
	errName := errors.New("name is reserved")
	var name, port, region string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name).Validate(func(string) error { return errName }),
		NewInput().Key("port").Value(&port).Validate(func(s string) error { return &portError{s} }),
		NewInput().Key("region").Value(&region),
	))

	err := New(form,
		WithNonInteractive(Always),
		WithEnvPrefix("APP"),
		WithAnswers(map[string]any{"name": "admin", "port": "x"}),
	).Run()
	if err == nil {
		t.Fatal("expected error")
	}
	want := `field "name": name is reserved
field "port": bad port x
missing required answers for:
  --region (env: APP_REGION)`
	if err.Error() != want {
		t.Errorf("expected:\n%s\ngot:\n%s", want, err.Error())
	}
	if !errors.Is(err, errName) {
		t.Errorf("expected errors.Is to reach name validator error")
	}
	var pe *portError
	if !errors.As(err, &pe) || pe.port != "x" {
		t.Errorf("expected errors.As to reach port validator error, got %v", pe)
	}
	if name != "" || port != "" {
		t.Errorf("expected invalid answers not written, got name=%q port=%q", name, port)
	}
}

func TestRunner_GroupWalk(t *testing.T) {
	cases := []struct {
		name           string
		answers        map[string]any
		want           string
		laterEvaluated bool
	}{
		{
			name:    "invalid answer stops the walk",
			answers: map[string]any{"first": "bad", "extra": "x"},
			want:    `field "first": bad first`,
		},
		{
			name:    "missing answers keep walking",
			answers: map[string]any{},
			want: "missing required answers for:\n" +
				"  --first  (env: APP_FIRST)\n" +
				"  --extra  (env: APP_EXTRA)\n" +
				"  --second (env: APP_SECOND)\n" +
				"  --third  (env: APP_THIRD)",
			laterEvaluated: true,
		},
		{
			name:    "invalid answer reports missing keys collected so far",
			answers: map[string]any{"first": "bad"},
			want: "field \"first\": bad first\n" +
				"missing required answers for:\n" +
				"  --extra (env: APP_EXTRA)",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var first, extra, second, third string
			hideCalls, optionCalls := 0, 0
			form := NewForm(
				NewGroup(
					NewInput().Key("first").Value(&first).Validate(func(s string) error {
						if s == "bad" {
							return errors.New("bad first")
						}
						return nil
					}),
					NewInput().Key("extra").Value(&extra),
				),
				NewGroup(
					NewSelect[string]().Key("second").Value(&second).
						OptionsFunc(func() []huh.Option[string] {
							optionCalls++
							return huh.NewOptions("a")
						}, &first),
					NewInput().Key("third").Value(&third),
				).WithHideFunc(func() bool {
					hideCalls++
					return false
				}),
			)
			hideCalls = 0 // huh evaluates the hide func while building the form.

			err := New(form,
				WithNonInteractive(Always),
				WithEnvPrefix("APP"),
				WithAnswers(tc.answers),
			).Run()
			if err == nil {
				t.Fatal("expected error")
			}
			if err.Error() != tc.want {
				t.Errorf("expected:\n%s\ngot:\n%s", tc.want, err.Error())
			}
			evaluated := hideCalls > 0 || optionCalls > 0
			if evaluated != tc.laterEvaluated {
				t.Errorf("expected later group evaluated=%v, got hide=%d options=%d", tc.laterEvaluated, hideCalls, optionCalls)
			}
		})
	}
}

func TestRunner_KeylessFieldJoinsGroupErrors(t *testing.T) {
	var name string
	form := NewForm(NewGroup(
		NewInput().Key("name").Value(&name).Validate(func(string) error { return errors.New("bad name") }),
		NewInput(),
	))
	err := New(form, WithNonInteractive(Always), WithAnswers(map[string]any{"name": "x"})).Run()
	if err == nil {
		t.Fatal("expected error")
	}
	for _, want := range []string{`field "name": bad name`, "required field 2 in group 1 has no Key() set"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected %q in error, got %q", want, err.Error())
		}
	}
}

func TestRunner_SingleErrorKeepsFormat(t *testing.T) {
	var name string
	form := NewForm(NewGroup(NewInput().Key("name").Value(&name).
		Validate(func(string) error { return errors.New("bad name") })))
	err := New(form, WithNonInteractive(Always), WithAnswers(map[string]any{"name": "x"})).Run()
	if err == nil || err.Error() != `field "name": bad name` {
		t.Errorf("expected single field error, got %v", err)
	}
}
