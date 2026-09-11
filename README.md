# volatile-downloads

Point `~/Downloads` at `/tmp/Downloads` at login so browser downloads stay
disposable. Replaces the 2014 Intel Automator applet (`Download-Workflow.app`).

## Install

On this Mac, chezmoi builds from the checkout and copies the binary to
`~/.local/bin` and `~/go/bin`, then loads the LaunchAgent.

On a machine without the checkout, chezmoi downloads the pinned GitHub
Release asset (`volatile-downloads-darwin-arm64` or `-amd64`).

Manual:

```bash
mise run install
```

Or:

```bash
go install github.com/zachbornheimer/volatile-downloads/cmd/volatile-downloads@v0.1.0
```

## What it does

1. Ensures `/tmp/Downloads` exists (mode 755, owned by you — no sudo).
2. If `~/Downloads` already points there, exits.
3. If `~/Downloads` is a real folder, moves its contents into `/tmp/Downloads`
   then replaces the folder with the symlink.
4. It will not delete a real Downloads folder, chmod 777, or kill Dock.

`/tmp` on macOS is not a ramdisk. Files usually survive reboot until the
system's periodic tmp cleanup.

## Flags

```
volatile-downloads --version
volatile-downloads --target /tmp/Downloads --link ~/Downloads
```

Env overrides: `VOLATILE_DOWNLOADS_TARGET`, `VOLATILE_DOWNLOADS_LINK`.
