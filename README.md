# huhx

**CLIs built on [huh](https://github.com/charmbracelet/huh) break in CI,
scripts, and agent-driven workflows.** huh has no way to skip the TUI and
accept answers programmatically, so every team that hits this hand-rolls
the same workaround: parallel flag handling, TTY detection, duplicated
validation logic, drifting code paths.

**huhx fixes this.** Build your form once. It runs as a beautiful TUI on
a terminal and accepts CLI flags, environment variables, or YAML/JSON
answer files everywhere else — CI pipelines, shell scripts, automated
tooling.

**Driving CLIs from agents.** AI agents and orchestrators can call
`WithAnswers(map[string]any{...})` to drive any huhx form in-process
without a terminal, reusing every validator the form already enforces.
No separate "headless mode" to keep in sync with the interactive one —
huhx is the same form.

The wrapper mirrors huh's full chainable API; existing huh code ports by
changing import paths.

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
                ).WithHideFunc(func() bool { return environment != "prod" }),
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
    flags.StringArray("answer", nil, "additional key=val answers")
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
3. `--answer key=val` entries from a `StringArray` flag named `answer`.
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

## Static vs dynamic options

`Select` and `MultiSelect` support both:

- `Options(opts...)` — static list captured at construction time.
- `OptionsFunc(f, bindings)` — dynamic provider re-evaluated lazily at
  injection time. Useful when the available choices depend on an
  earlier field's value (e.g. State depending on Country).

When using `OptionsFunc`, the dependent field must live in a **later
group** than its source field. The non-interactive runner walks groups
in order and writes each field's value before later groups resolve, so
closures capturing earlier-field pointers see the right values. This is
the same constraint huh's interactive `bindings` machinery already
enforces.

Calling `Options(...)` clears any prior `OptionsFunc(...)` and vice
versa — last setter wins.

## Conditional groups

`Group.WithHide(bool)` and `Group.WithHideFunc(func() bool)` skip the
group in non-interactive mode and hide it in interactive mode. Both
mirror huh's API exactly — `WithHide` takes a static bool, `WithHideFunc`
takes a predicate re-evaluated at run time.

## Missing-answer error

```
missing required answers for:
  --name        (env: DEPLOY_NAME)
  --environment (env: DEPLOY_ENVIRONMENT)
```

## Migrating from huh

Migration is mostly mechanical — one decision per field.

### 1. Import + constructors

```go
// before
import "charm.land/huh/v2"

form := huh.NewForm(
    huh.NewGroup(
        huh.NewInput().Title("Name").Value(&name),
    ),
)
if err := form.Run(); err != nil { ... }
```

```go
// after
import (
    "charm.land/huh/v2"
    "github.com/cabljac/huhx"
)

form := huhx.NewForm(
    huhx.NewGroup(
        huhx.NewInput().Key("name").Title("Name").Value(&name),
    ),
)
runner := huhx.New(form,
    huhx.WithEnvPrefix("MYAPP"),
    huhx.WithCobraFlags(cmd), // if cobra is wired
)
if err := runner.Run(); err != nil { ... }
```

Keep `huh.NewOption`, `huh.Option[T]`, `huh.Accessor[T]`, theme types,
etc. — huhx reuses huh's types unchanged.

### 2. The Key decision

Every field that should be drivable non-interactively needs a `.Key(k)`.
The key becomes the CLI flag name (`--my-key`), the environment variable
suffix (`PREFIX_MY_KEY`), and the answer file key. Pick keys that read
well as flags — lowercase, hyphen-separated.

```go
huhx.NewInput().Key("repo-name").Title("Repository name").Value(&repoName)
// non-interactive: --answer repo-name=...   MYAPP_REPO_NAME=...
```

Fields without `.Key()` still work in interactive mode — huhx forwards
them to huh as normal. Non-interactive behavior:

- **Required keyless field** → runner errors with
  `required field at group N, position M has no Key() set; call .Key("...") on it to enable non-interactive mode`.
  Run the binary once in non-interactive mode, see which field needs a
  key, add it, repeat.
- **Optional keyless field** (`.Optional()`) → silently skipped in
  non-interactive mode.

That's the whole migration loop. The rest is search-and-replace.

### 3. WithCobraFlags wiring (cobra users)

Register matching flags on your cobra command so huhx can read named
flag values:

```go
cmd.Flags().String("repo-name", "", "")
cmd.Flags().StringArray("answer", nil, "additional answers in key=val form")
cmd.Flags().String("answer-file", "", "path to YAML/JSON answer file")
cmd.Flags().Bool("non-interactive", false, "force non-interactive mode")
```

Then pass `huhx.WithCobraFlags(cmd)` to the runner.
