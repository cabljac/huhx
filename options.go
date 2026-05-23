package huhless

import "github.com/spf13/cobra"

// Option configures a Runner.
type Option func(*Runner)

// WithEnvPrefix sets the prefix used when looking up answers in environment
// variables. A field with key "name" and prefix "MYCLI" is looked up as
// MYCLI_NAME.
func WithEnvPrefix(prefix string) Option {
	return func(r *Runner) { r.envPrefix = prefix }
}

// WithAnswers supplies answers programmatically. Values are stringified
// via fmt.Sprintf("%v", v) before being handed to the field's setter.
// This has the highest precedence among answer sources.
func WithAnswers(m map[string]any) Option {
	return func(r *Runner) { r.answers = m }
}

// WithAnswerFile loads answers from a YAML or JSON file. YAML is a superset
// of JSON so a single decoder handles both.
func WithAnswerFile(path string) Option {
	return func(r *Runner) { r.answerFile = path }
}

// WithCobraFlags wires the runner to a cobra command. The runner uses
// the command for three things:
//
//  1. Looking up each field's key as a named flag (e.g. --name) and
//     using its value when the flag has been explicitly set.
//  2. Reading --answer key=val pairs from a StringSlice flag named
//     "answer" if present.
//  3. Honouring a --non-interactive bool flag if present and set, as
//     one of the AutoDetect inputs.
func WithCobraFlags(cmd *cobra.Command) Option {
	return func(r *Runner) { r.cobraCmd = cmd }
}

// WithNonInteractive overrides the mode selection. Default is AutoDetect.
func WithNonInteractive(mode Mode) Option {
	return func(r *Runner) { r.mode = mode }
}
