# Contributing to huhx

Thanks for considering a contribution. This document covers the bits that
matter day-to-day; if something is missing, open an issue.

## Scope

huhx is a thin wrapper over [charmbracelet/huh](https://github.com/charmbracelet/huh).
Changes generally fall into one of these buckets:

- **Bug fixes** to the non-interactive runner, source precedence, or
  field setter behavior.
- **API parity** with huh — new chainable methods on the wrappers should
  forward to the inner field and return the wrapper.
- **New non-interactive surfaces** — additional answer sources, new
  resolution rules.

Out of scope:

- Forking huh's rendering behavior. If a TUI change is needed, file it
  upstream against huh.
- huhx-only fields that aren't backed by a huh field type.

## Development

```bash
git clone https://github.com/cabljac/huhx.git
cd huhx
go test ./...
```

`go vet ./...` must be clean. `gofmt -l .` must produce no output.

## Testing

- Use `test` not `it` for subtest function names.
- Each test runs the runner in `Always` non-interactive mode unless the
  test is specifically about mode detection.
- For new chainable forwarders, add both:
  1. A chainability test that calls every new method in one chain and
     asserts the wrapper is non-nil.
  2. Where behavior is observable end-to-end (e.g. `Accessor`), an
     integration test that drives a runner non-interactively and
     asserts the bound value.
- Coverage target: 90%+ statement coverage. Pure delegation setters
  don't have to hit 100%; behavioral paths should.
- The end-to-end suite (`e2e_test.go`) builds the deploy example once
  via `TestMain` and exercises it as a subprocess. Anything that
  changes observable subprocess behavior needs an e2e test.

Run the e2e suite specifically:

```bash
go test -run TestE2E ./...
```

## Commits

Conventional Commits style. Subject line ≤72 chars. Body explains the
**why** more than the **what**.

Examples from this repo:

```
feat(input): full huh.Input parity
fix: read --answer as StringArray not StringSlice
test(e2e): cover Text and MultiSelect via the deploy example
docs: lead with the pain in README opening
```

Atomic commits — one logical change per commit. If you touch six files
for one rename and three files for one feature, that is two commits.

Do not add Claude (or any AI tooling) as a co-author. Attribution stays
human.

## Pull requests

- Branch off `main`.
- Open the PR from your fork; CI must be green.
- Link the issue if there is one.
- For API additions, include the matching test from the Testing section
  above and a one-line README note if the addition is user-visible.

## Code style

- Default to no comments. Add one only when the *why* is non-obvious.
- Don't write doc comments that restate the method name.
- Don't add error handling for impossible cases. Trust internal
  guarantees; validate only at system boundaries.
- Don't use `any` or `interface{}` unless required by an external
  signature.
