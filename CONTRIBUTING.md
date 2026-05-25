# Contributing to go-openfga

Thanks for your interest in improving go-openfga! This document covers the
development workflow, the conventions the project follows, and how to get your
change merged.

By participating you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md).

## Getting started

```bash
git clone https://github.com/sergiught/go-openfga
cd go-openfga
make test        # unit tests with the race detector
make lint        # golangci-lint (installs into ./bin on first run)
```

Run `make help` to see every available target. The first run of `make lint`,
`make lint-commits`, or `make vuln` installs the matching tool (SHA-pinned) into
`./bin` — nothing is installed system-wide.

### Optional: git hooks

If you have [pre-commit](https://pre-commit.com) installed, wire up the local
hooks so formatting, linting, and commit-message checks run automatically:

```bash
make pre-commit
```

## Project layout

| Path | What lives there |
| --- | --- |
| `openfga/` | The client library — the only package consumers import. |
| `test/integration/` | A separate Go module: a [godog](https://github.com/cucumber/godog) acceptance suite that runs against a real OpenFGA server via [testcontainers](https://golang.testcontainers.org). Its test-only dependencies never enter the client's public dependency graph. |

## Making a change

1. **Open an issue first** for anything beyond a small fix, so we can agree on the
   approach before you invest time.
2. **Keep PRs focused.** One logical change per PR; split refactors from behavior
   changes.
3. **Add tests.** Unit tests live next to the code in `openfga/*_test.go`. Behavior
   that needs a live server belongs in `test/integration/features/*.feature` with the
   step bindings in `test/integration`.
4. **Document exported symbols.** Every exported identifier needs a godoc comment;
   `golangci-lint` enforces this.
5. **Run the checks locally** before pushing:

   ```bash
   make lint
   make test
   make integration   # requires Docker; spins up OpenFGA via testcontainers
   ```

## Commit and PR conventions

This project uses [Conventional Commits](https://www.conventionalcommits.org/).
Commit subjects (and PR titles, which become the squash-merge subject) must look
like:

```
type(optional-scope): short imperative summary
```

Allowed types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`,
`build`, `ci`, `chore`, `revert`. The PR-title check in CI rejects anything else.
You can validate a message locally with `make lint-commits` (it reads
`.git/COMMIT_EDITMSG`).

Examples:

```
feat: add StreamedListObjects NDJSON iterator
fix: stop ChangesAll once the changes feed is drained
docs: document the private-key JWT auth option
```

## Continuous integration

Every PR runs:

- **lint** — `golangci-lint` plus a `go.mod`/`go.sum` drift check.
- **test** — unit tests with `-race` and coverage, across the supported Go floor and
  the current stable release.
- **integration** — the testcontainers + godog suite against a real OpenFGA server.
- **govulncheck** — the module graph scanned for known vulnerabilities.
- **codeql** — static security analysis.
- **pr-title** — Conventional Commits check on the PR title.

All of these must be green before a PR can merge.

## Reporting bugs and requesting features

Use the [issue templates](https://github.com/sergiught/go-openfga/issues/new/choose).
For security vulnerabilities, **do not open a public issue** — follow the
[security policy](SECURITY.md) instead.

## License

By contributing, you agree that your contributions will be licensed under the
[MIT License](LICENSE) that covers the project.
