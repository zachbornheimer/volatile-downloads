# 1. Replace the Automator login item with a Go binary

Date: 2026-09-11

## Status

Accepted

## Context

Download-Workflow.app (2014) is an Intel-only Automator stub that ran a
C helper to point `~/Downloads` at `/tmp/Downloads`. macOS 26 warns that
Intel components will stop working. The C helper also `rm -R`'d Downloads,
chmod 777, and killed Dock.

## Decision

Ship a signed arm64 Go binary. Chezmoi installs it to `~/.local/bin` and
`~/go/bin`. A LaunchAgent runs it at login. GitHub Actions attach
darwin/arm64 and darwin/amd64 binaries to version tags so a machine
without a checkout can `gh release download`.

## Consequences

No Rosetta. No sudo. No Automator. A real Downloads folder is merged,
never deleted. Future macOS tmp-cleanup policy still governs how long
files in `/tmp/Downloads` live.
