# Agent guide

This file is for AI agents (Claude, Codex, Aider, etc.) working on huhx.
It is shorter and more directive than the human-facing CONTRIBUTING.md.
If you are a human, read CONTRIBUTING.md instead.

## What huhx is

A thin wrapper around [charmbracelet/huh](https://github.com/charmbracelet/huh)
that adds a non-interactive Runner. Forms are built with huhx builders;
the Runner drives them interactively on a TTY (delegating to huh) or
non-interactively by walking groups in order and resolving each field's
answer from configured sources.

Module path: `github.com/cabljac/huhx`. Go version: 1.25.

## Repo map

```
field.go                    internal field interface
input.go text.go            string fields
confirm.go                  bool field
select.go multiselect.go    generic [T comparable] fields
group.go form.go            composition wrappers
mode.go                     Mode enum (AutoDetect/Always/Never)
options.go                  Option type + With* setters
sources.go                  answer-file (YAML/JSON), --answer parsing, env key formatting
runner.go                   Runner.Run, walk, resolve, missing-err
*_test.go                   unit + integration tests, all in package huhx
e2e_test.go                 subprocess tests in package huhx_test
examples/deploy/main.go     cobra-wired example used by e2e
```

## Conventions

- **Tests**: use `test` not `it`. Subtests via `t.Run`. Each forwarder
  added on a field wrapper must come with a chainability test and, if
  the forwarder is observable in non-interactive mode, an integration
  test.
- **Comments**: default to none. Add one only when the *why* is non-
  obvious (a hidden constraint, a workaround, surprising behavior).
  Never restate what well-named code already says.
- **Errors**: wrap with `%w` and field/source context. Validators are
  surfaced via the runner as `field %q: %w`.
- **No `any`/`interface{}`** unless required by an external signature.
- **Atomic commits**: one logical change per commit. Conventional
  Commits style. Subject ≤72 chars. Body explains the *why*.
- **No co-author attribution for AI tools**. Don't add Claude (or any
  agent) as a Co-Authored-By.

## Field wrapper pattern

Every field type follows the same shape:

```go
type X struct {
    inner       *huh.X
    k           string
    value       *T
    accessor    huh.Accessor[T]
    validate    func(T) error
    optional    bool
    // type-specific extras: options, optionsFunc, etc.
}
```

When adding a new chainable forwarder, mirror huh's exact signature and
return the huhx wrapper. For state-affecting calls (Accessor,
OptionsFunc) also store the state on the wrapper so the non-interactive
`set()` can use it.

## Adding a new chainable

1. Find the method's signature in the huh source.
2. Add the method to the matching huhx file with `inner.<Method>(args...)`
   and `return wrapper`.
3. Add a one-line doc comment.
4. Add the method to the chainability test in `<type>_test.go`.

## Adding a new answer source

1. Plumb the option through `options.go` (`WithFoo`) and store on
   `Runner`.
2. Add the resolution step to `Runner.resolve` at the right precedence
   layer (see README "Answer source precedence").
3. Add an integration test in `runner_test.go` proving it wins over the
   layer below and loses to the layer above.

## Running things

```bash
go test ./...                       # all tests, ~46
go test -run TestE2E ./...          # subprocess suite
go test -coverprofile=/tmp/c.cov ./... && go tool cover -func=/tmp/c.cov
go vet ./...
gofmt -l .                          # must be empty
go build ./...
```

## Out of scope

- Forking huh's TUI rendering. File upstream against huh.
- huhx-only fields without a huh backing type.
- Changes that diverge huhx's API from huh's. Parity is a non-goal-
  breaker — when in doubt, mirror huh.

## Don't

- Don't add `// removed X` stubs for things you delete. Just delete them.
- Don't write "this method ..." doc comments that restate the name.
- Don't add error handling for impossible cases. Trust internal
  guarantees.
- Don't bypass `go vet` or `gofmt`. Fix the issue.
- Don't push to `main` without atomic commits.
- Don't add yourself as Co-Authored-By.
