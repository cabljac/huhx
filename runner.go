package huhless

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
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
// from the configured sources, returning a descriptive error if any
// required answer is missing or any validator fails.
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
	for _, g := range r.form.groups {
		if g.hide != nil && g.hide() {
			continue
		}
		for _, f := range g.fields {
			ans, ok := r.resolve(f.key(), fileAns, cliAns)
			if !ok {
				if f.required() {
					missing = append(missing, f.key())
				}
				continue
			}
			if err := f.set(ans); err != nil {
				return fmt.Errorf("field %q: %w", f.key(), err)
			}
		}
	}
	if len(missing) > 0 {
		return r.missingErr(missing)
	}
	return nil
}

func (r *Runner) resolve(key string, fileAns, cliAns map[string]string) (string, bool) {
	if v, ok := r.answers[key]; ok && v != nil {
		return fmt.Sprintf("%v", v), true
	}
	if r.cobraCmd != nil {
		if f := r.cobraCmd.Flags().Lookup(key); f != nil && f.Changed {
			return f.Value.String(), true
		}
	}
	if v, ok := cliAns[key]; ok {
		return v, true
	}
	if v, ok := fileAns[key]; ok {
		return v, true
	}
	if r.envPrefix != "" {
		if v, ok := os.LookupEnv(envKey(r.envPrefix, key)); ok {
			return v, true
		}
	}
	return "", false
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
