# rr-policy-guards

Mandatory Rational Reserve policy hooks for Sovereign projects.

## Guard set

Six compiled Go guards run as Claude Code `PreToolUse` hooks:

- `rr-emoji-guard` validates `Write`, `Edit`, and `MultiEdit` content. It accepts valid UTF-8 language and mathematical text while blocking invalid UTF-8, emoji, and prohibited decorative Unicode.
- `rr-bash-guard` validates every Bash request. It allows quoted variable expansion and blocks unsafe unquoted `$VAR` or `${VAR}` expansion.
- `rr-brew-guard` blocks dangerous Homebrew flags, URL installs, and untrusted taps.
- `rr-tofu-guard` blocks direct `tofu apply`, `destroy`, and `import` requests. Approved lifecycle scripts remain the required execution path and do not use a guard bypass.
- `rr-commit-guard` scans staged paths and inline commit messages. It detects commits in compound shell commands and fails closed when directory-changing shell state makes the repository ambiguous.
- `rr-verify-guard` detects `git commit` and `git push` in executable shell segments, including pipelines, subshells, and command substitutions. It resolves an explicit `git -C PATH` target, runs local CI-equivalent checks, and blocks degraded verification.

There are no guard bypass variables or commit-message waiver tags.

## Publication gate

Commit and push must be separate requests. Before every push, `rr-verify-guard` requires:

1. One unambiguous Git repository.
2. A clean staged, unstaged, and untracked state.
3. Applicable lint, type-check, test, build, workflow, runner, and action checks.
4. A fresh Semgrep scan (`p/security-audit`).
5. A fresh Gitleaks scan.
6. Fresh dependency SCA for every discovered module root:
   - `yarn.lock` roots: `yarn npm audit --all --recursive --severity high --no-deprecations`
   - `package-lock.json` roots (without yarn.lock): `npm audit --audit-level=high`
   - `go.mod` roots: `govulncheck ./...`

Push never uses a pipeline-cache hit as a waiver. A missing scanner, skipped check, degraded external check, or failed check blocks publication.

Use the Bash tool working directory or one explicit `git -C PATH` target. Commands such as `cd PATH && git push` are rejected because a pretool hook cannot safely inherit the shell's future directory state.

## Layout

```text
plugins/rr-policy-guards/
|-- plugin.json
|-- hooks/hooks.json
|-- git-hooks/
|   |-- pre-commit
|   `-- commit-msg
|-- scripts/install-git-hooks.sh
|-- tools/
|   |-- emoji-guard/
|   |-- bash-guard/
|   |-- brew-guard/
|   |-- tofu-guard/
|   |-- commit-guard/
|   `-- verify-guard/
`-- bin/
    |-- rr-emoji-guard
    |-- rr-bash-guard
    |-- rr-brew-guard
    |-- rr-tofu-guard
    |-- rr-commit-guard
    `-- rr-verify-guard
```

The binaries are generated and ignored by Git.

## Build and test

Run each command from its guard directory:

```sh
go test ./...
go vet ./...
go build -o ../../bin/rr-<name>-guard .
```

The implementation uses the Go standard library. The verification guard invokes approved external project tools and requires Semgrep, Gitleaks, and dependency SCA (`yarn`/`npm` audit and `govulncheck`) for push.

## Hook configuration

The packaged configuration is `hooks/hooks.json`. Active Sovereign configuration is mirrored in:

- `~/Projects/Sovereign/Structure/hooks/hooks.json`
- `~/Projects/Sovereign/.claude/settings.json`
- `~/.claude/settings.json`

Commit and verification guards are unconditional members of the Bash matcher and perform fast internal target filtering. Prefix-only hook `if` expressions are intentionally absent so compound commands cannot evade enforcement.

Hook and binary changes require a new Claude Code session before they become active.

## Audit logs

Default logs are mode-0600 JSONL files under `~/.rational-reserve/logs/`. Override paths remain available through each guard's `RR_<NAME>_GUARD_AUDIT_LOG` variable; this changes only the audit destination and never changes a decision.

The Bash and verification logs rotate at 8 MiB with three numbered backups. Their byte limits can be tuned with `RR_BASH_GUARD_AUDIT_MAX_BYTES` and `RR_VERIFY_GUARD_AUDIT_MAX_BYTES`. The Structure-owned posttool guard applies the same default through `RR_POSTTOOL_GUARD_AUDIT_MAX_BYTES`.

Audit records omit secrets and do not create a policy waiver.

## Git commit hooks

`scripts/install-git-hooks.sh [REPO_PATH]` installs `pre-commit` and `commit-msg` hooks that delegate to `rr-commit-guard`. Existing unrelated hooks are backed up before replacement. These Git hooks supplement the agent pretool hooks; they do not replace the publication gate.

## License

MIT.
