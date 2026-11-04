package huhx

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// envTruthy reports whether an environment variable's value should be
// treated as true. Accepts "1", "true"/"TRUE" (and friends via
// strconv.ParseBool), plus "yes"/"YES" as an explicit extension.
func envTruthy(name string) bool {
	v := os.Getenv(name)
	if v == "" {
		return false
	}
	if strings.EqualFold(v, "yes") {
		return true
	}
	b, err := strconv.ParseBool(v)
	return err == nil && b
}

// Runner drives a Form either interactively (delegating to huh) or
// non-interactively (walking groups and injecting answers from configured
// sources).
type Runner struct {
	form *Form
	mode Mode

	envPrefix  string
	answers    map[string]any
	answerFile string
	cobraCmd   *cobra.Command
}

// New returns a Runner. Default mode is AutoDetect.
func New(form *Form, opts ...Option) *Runner {
	r := &Runner{form: form, mode: AutoDetect}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Run executes the form. In interactive mode it delegates to huh.Form.Run().
// In non-interactive mode it walks groups and resolves each field's answer
// from the configured sources. Missing answers are collected across all
// groups. A group with an invalid value stops the walk at its end. The
// returned error joins that group's field errors and the missing answers
// found so far, so errors.Is and errors.As reach each validator error.
func (r *Runner) Run() error {
	if r.isNonInteractive() {
		return r.runNonInteractive()
	}
	return r.form.inner.Run()
}

// isNonInteractive reports whether the runner should skip the bubble tea
// loop and resolve answers from configured sources instead.
func (r *Runner) isNonInteractive() bool {
	switch r.mode {
	case Always:
		return true
	case Never:
		return false
	}
	if envTruthy("NON_INTERACTIVE") || envTruthy("CI") {
		return true
	}
	if !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd()) {
		return true
	}
	if r.cobraCmd != nil {
		if f := r.cobraCmd.Flags().Lookup("non-interactive"); f != nil && f.Changed && f.Value.String() == "true" {
			return true
		}
	}
	return false
}

func (r *Runner) runNonInteractive() error {
	fileAns, err := loadAnswerFile(r.answerFile)
	if err != nil {
		return err
	}
	cliAns, err := cobraAnswerPairs(r.cobraCmd)
	if err != nil {
		return err
	}

	var missing []string
	for gi, g := range r.form.groups {
		if g.hide != nil && g.hide() {
			continue
		}
		errs, groupMissing := r.runGroup(gi, g, fileAns, cliAns)
		missing = append(missing, groupMissing...)
		// A missing answer writes nothing, so later groups only need
		// protecting from invalid values.
		if len(errs) > 0 {
			return r.joinErrs(errs, missing)
		}
	}
	return r.joinErrs(nil, missing)
}

// runGroup answers every field in g. It returns the invalid-value errors
// and the keys of required fields that have no answer.
func (r *Runner) runGroup(gi int, g *Group, fileAns map[string]any, cliAns map[string]string) (errs []error, missing []string) {
	for fi, f := range g.fields {
		if f.key() == "" {
			if f.required() {
				errs = append(errs, fmt.Errorf("required field %d in group %d has no Key() set; call .Key(\"...\") on it to enable non-interactive mode", fi+1, gi+1))
			}
			continue
		}
		ans, ok := r.resolve(f.key(), fileAns, cliAns)
		if !ok {
			hasDefault, err := f.checkDefault()
			if err != nil {
				errs = append(errs, fmt.Errorf("field %q: %w", f.key(), err))
			} else if !hasDefault && f.required() {
				missing = append(missing, f.key())
			}
			continue
		}
		if err := f.set(ans); err != nil {
			errs = append(errs, fmt.Errorf("field %q: %w", f.key(), err))
		}
	}
	return errs, missing
}

// joinErrs returns errs and the missing-answers block as one error, or a
// lone error unwrapped so its text and type stay as they were.
func (r *Runner) joinErrs(errs []error, missing []string) error {
	if len(missing) > 0 {
		errs = append(errs, r.missingErr(missing))
	}
	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	}
	return errors.Join(errs...)
}

func (r *Runner) resolve(key string, fileAns map[string]any, cliAns map[string]string) (any, bool) {
	if v, ok := r.answers[key]; ok && v != nil {
		return v, true
	}
	if r.cobraCmd != nil {
		if f := r.cobraCmd.Flags().Lookup(key); f != nil && f.Changed {
			// String() on a slice flag renders "[a,b]", which no field can parse.
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				return sv.GetSlice(), true
			}
			return f.Value.String(), true
		}
	}
	if v, ok := cliAns[key]; ok {
		return v, true
	}
	if v, ok := fileAns[key]; ok && v != nil {
		return v, true
	}
	if r.envPrefix != "" {
		if v, ok := os.LookupEnv(envKey(r.envPrefix, key)); ok {
			return v, true
		}
	}
	return nil, false
}

func (r *Runner) missingErr(missing []string) error {
	width := 0
	for _, k := range missing {
		if w := len(k) + 2; w > width {
			width = w
		}
	}
	var b strings.Builder
	b.WriteString("missing required answers for:\n")
	for _, k := range missing {
		fmt.Fprintf(&b, "  %-*s", width, "--"+k)
		if env := envKey(r.envPrefix, k); env != "" {
			fmt.Fprintf(&b, " (env: %s)", env)
		}
		b.WriteByte('\n')
	}
	return errors.New(strings.TrimRight(b.String(), "\n"))
}
