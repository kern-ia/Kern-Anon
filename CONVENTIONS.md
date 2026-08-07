# CONVENTIONS.md — kern-anon

Local authority for this repo, as announced by the org-wide
[CONTRIBUTING.md](https://github.com/kern-ia/.github/blob/main/CONTRIBUTING.md). The rules
shared by all `kern-ia` repos are restated below; the "Specifics" sections cover what belongs
only to `kern-anon`.

## Language

Code, identifiers, and comments are written in English — no exceptions. This applies to
source files, docstrings, commit diffs, and test names. Internal documentation such as this
file, `README.md`, or `CLAUDE.md` stays in whatever language the team works in day to day.

> **Conflict to resolve**: `.golangci.yml` currently disables `misspell` with the comment
> "misspell is disabled: it assumes English comments and flags the project's French
> (`contextuel`, `collecte`...) as errors." That comment describes the codebase as it stands
> today — a fair amount of existing comments are in French. Under this new rule that
> justification no longer holds going forward: new comments must be English. Existing French
> comments are legacy, not required to be rewritten in one pass, but `misspell` should
> eventually be re-enabled once the backlog is cleared rather than left off indefinitely.

## Branches

- `main`: stable branch, always deployable. Protected — no direct pushes.
- `dev`: integration branch. Protected — no direct pushes.
- Working branches: `feature/<slug>`, `fix/<slug>`, `chore/<slug>`, `docs/<slug>`, `test/<slug>`.
- Any change to `main` or `dev` goes through a Pull Request. This repo's `CLAUDE.md` already
  says "never commit directly to main/dev" — that rule is not technically enforced today (no
  branch protection enabled on GitHub) and is not followed via PR either (0 PRs opened so
  far, everything comes from locally pushed merges). To fix: enable branch protection AND go
  through a PR, even solo.
- Merging into `dev`: `--no-ff` ONLY if tests are green and E2E is done (rule already in
  force, keep it).

## Commits

Conventional Commits: `feat:`, `fix:`, `test:`, `docs:`, `chore:`... (already respected). No
tool signature (`Co-Authored-By`, `Claude-Session`, or equivalent trailer) in commit
messages — the git author is enough.

## Pull Requests

- One subject per PR, linked to the issue or RFC it resolves.
- PR template inherited from `kern-ia/.github`.
- States the semver impact.
- No real personal data — critical here since the repo's whole purpose is handling synthetic
  PII by construction.

## Method (already in force, keep it)

- **Strict TDD**: tests written before the code, `go test ./...` green before every commit.
- Test cases derived from the oracle corpus (`internal/testdata/oracle.jsonl` /
  `anonymize.jsonl`).
- Business logic stays pure in the domain packages (`recognizer`, `analyzer`, `anonymizer`...),
  no I/O in them.
- Offsets expressed in **runes**, never bytes — tested with accents/emoji on every recognizer.
- `nlp/onnx` stays opt-in behind the `onnx` build tag (cgo); without the tag, 100% pure Go.

## Style and lint

`.golangci.yml` — `version: 2`, `linters.default: standard` + `revive`, `gocritic`,
`prealloc`. `misspell` is currently disabled for the reason noted above under "Language" —
revisit once new comments are consistently English. Any future extension of the linter set
must carry the same kind of justifying comment as the existing ones.
`max-issues-per-linter: 0`, `max-same-issues: 0` — nothing is hidden.

## Tests / CI

- `go build ./...`, `go build -tags onnx ./...`, `go test -race -cover ./...`, lint via
  `golangci-lint`. Already in place in `.github/workflows/ci.yml` — keep it as the reference
  for other Go repos in the org that don't have CI yet (`kern-orch`).

## Go module

> **Current gap — the most visible one in the audit**: `go.mod` still declares
> `module github.com/YoLaub/PresidioGo`, a leftover from the name before the rename to
> `kern-anon`. This path matches neither the repo name (`kern-anon`) nor the organization
> (`kern-ia`). To fix: rename the module (`go mod edit -module` + update all internal
> imports), coordinated with the org-level decision on the `github.com/kern-ia/...` path (see
> the global report).

## Documentation

- `README.md`, `LICENSE` at the root (both present — a good example for the other repos).
- `CLAUDE.md` — agent context.
- OKF index under `docs/index/`: one sheet per completed feature (YAML header: id, feature,
  branch, status, files, tests, decisions; body ≤ 15 lines). Pitfalls logged in
  `docs/index/retro.md` at the moment they bite.
- No `CHANGELOG.md`: release notes live in the annotated tag (org convention).

## Security / privacy

See the org-inherited `SECURITY.md`. Privacy-boundary defects (un-pseudonymized PII reaching a
provider, verbatim content being logged) are treated with the same severity as a
code-execution bug — an explicit reminder for this repo in particular.
