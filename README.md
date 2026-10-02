# shell-mcp

<p align="center">
  <img src="assets/banner.svg" alt="shell-mcp — one command, exit code first" width="100%">
</p>

<p align="center">
  <a href="https://github.com/shotah/shell-mcp/actions/workflows/ci.yml"><img src="https://github.com/shotah/shell-mcp/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/shotah/shell-mcp/actions/workflows/release.yml"><img src="https://github.com/shotah/shell-mcp/actions/workflows/release.yml/badge.svg" alt="Release"></a>
  <a href="https://github.com/shotah/shell-mcp/actions/workflows/ci.yml"><img src="https://github.com/shotah/shell-mcp/raw/gh-pages/badges/coverage.svg" alt="Coverage"></a>
  <a href="https://pkg.go.dev/github.com/shotah/shell-mcp"><img src="https://pkg.go.dev/badge/github.com/shotah/shell-mcp.svg" alt="Go Reference"></a>
  <img src="https://img.shields.io/github/go-mod/go-version/shotah/shell-mcp" alt="Go version">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/shotah/shell-mcp" alt="License"></a>
</p>

One shell command over MCP stdio. Server id `shell`. The host calls `shell__command_run`.

The working directory starts at `--root`. The container, or that directory on a host unit, is the sandbox. Tests and builds are arguments to `command_run`, not extra tools.

Naming: [george docs/mcp-naming.md](https://github.com/shotah/george/blob/main/docs/mcp-naming.md).

## Run

```bash
go install github.com/shotah/shell-mcp@latest
shell-mcp --root /workspace
```

```toml
[[server]]
name    = "shell"
command = "shell-mcp"
args    = ["--root", "/workspace"]
force   = true
```

## Development

```bash
make lint
make test
make coverage   # fails below 70%
make release    # patch bump; BUMP=minor|major or TAG=v0.2.0
```

CI (`.github/workflows/ci.yml`) runs the same lint, tests, and `scripts/check-coverage.sh` at 70%. `make check` is the local equivalent. `make release` writes `VERSION`, tags `v*` and floating `latest`, and pushes so the Release workflow runs.
