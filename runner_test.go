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
		).WithHide(func() bool { return hidden }),
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
	if !strings.Contains(msg, "--all-regions") {
		t.Errorf("expected --all-regions listed, got %q", msg)
	}
	if !strings.Contains(msg, "(env: DEPLOY_ENVIRONMENT)") {
		t.Errorf("expected env hint for environment, got %q", msg)
	}
	if !strings.Contains(msg, "(env: DEPLOY_ALL_REGIONS)") {
		t.Errorf("expected env hint for all-regions, got %q", msg)
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
	cmd.Flags().StringSlice("answer", nil, "")
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
