# huhx

Thin builder layer on top of [charmbracelet/huh](https://github.com/charmbracelet/huh)
that adds non-interactive / headless execution.

Build a form once with the huhx builders. The runner drives it
interactively on a TTY and falls back to CLI flags, env vars, and answer
files in CI.

## Install

```bash
go get github.com/cabljac/huhx
```

## Quick start

```go
package main

import (
    "fmt"
    "os"

    "charm.land/huh/v2"
    "github.com/cabljac/huhx"
    "github.com/spf13/cobra"
)

func main() {
    var (
        name        string
        environment string
        allRegions  bool
    )

    cmd := &cobra.Command{
        Use: "deploy",
        RunE: func(cmd *cobra.Command, args []string) error {
            form := huhx.NewForm(
                huhx.NewGroup(
                    huhx.NewInput().Key("name").Title("App name").Value(&name),
                    huhx.NewSelect[string]().Key("environment").Title("Environment").
                        Options(
                            huh.NewOption("staging", "staging"),
                            huh.NewOption("prod", "prod"),
                        ).Value(&environment),
                ),
                huhx.NewGroup(
                    huhx.NewConfirm().Key("all-regions").Title("Deploy to all regions?").Value(&allRegions),
                ).WithHide(func() bool { return environment != "prod" }),
            )

            runner := huhx.New(form,
                huhx.WithEnvPrefix("DEPLOY"),
                huhx.WithCobraFlags(cmd),
            )
            return runner.Run()
        },
    }
    flags := cmd.Flags()
    flags.String("name", "", "")
    flags.String("environment", "", "")
    flags.Bool("all-regions", false, "")
    flags.StringSlice("answer", nil, "additional key=val answers")
    flags.Bool("non-interactive", false, "force non-interactive mode")

    if err := cmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}
```

## Running the deploy example

```bash
# interactive
go run ./examples/deploy

# non-interactive
CI=1 go run ./examples/deploy \
  --answer name=myapp \
  --answer environment=prod \
  --answer all-regions=true
```

## Mode selection

`huhx.AutoDetect` (default) picks non-interactive when any of:

1. `NON_INTERACTIVE=1` or `CI=1` is set.
2. stdin is not a TTY.
3. `--non-interactive` flag is present on the wired cobra command.

Otherwise the runner delegates to `huh.Form.Run()`.

Force the mode with `huhx.WithNonInteractive(huhx.Always | huhx.Never)`.

## Answer source precedence

When non-interactive, each field's answer is resolved in order:

1. `WithAnswers(map[string]any{...})` — programmatic injection.
2. Cobra named flag matching the field key (e.g. `--name`).
3. `--answer key=val` entries from a `StringSlice` flag named `answer`.
4. Answer file from `WithAnswerFile(path)` (YAML or JSON).
5. Environment variable `<PREFIX>_<KEY>` (with `WithEnvPrefix`).
6. Otherwise the field is reported as missing.

## Field types

| huhx | wraps |
|---|---|
| `Input` | `*huh.Input` |
| `Text` | `*huh.Text` |
| `Confirm` | `*huh.Confirm` |
| `Select[T]` | `*huh.Select[T]` |
| `MultiSelect[T]` | `*huh.MultiSelect[T]` |

Each wrapper mirrors huh's chainable API. `Validate(fn)` stores the
validator on the wrapper so it runs against headless-injected values
without going through huh internals.

`MultiSelect` accepts comma-separated answers (`a,b,c`).

`Confirm` parses with `strconv.ParseBool`.

## Conditional groups

`Group.WithHide(fn func() bool)` skips the group in non-interactive mode
and hides it in interactive mode when `fn()` returns true. Mirrors huh's
`WithHideFunc` naming.

## Missing-answer error

```
missing required answers for:
  --name        (env: DEPLOY_NAME)
  --environment (env: DEPLOY_ENVIRONMENT)
```
