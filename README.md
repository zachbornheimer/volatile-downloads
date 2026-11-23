# volatile-downloads

`~/Downloads` is a junk drawer. This points it at `/tmp/Downloads` at login so
browser downloads stay disposable and do not pile up in home.

You do **not** need a checkout under `~/Developer/Personal/`. Delete this
repo whenever you want. Chezmoi installs the binary from GitHub Releases.

## Install (the real path)

On macOS, `chezmoi apply` (or `chezmoi init --apply` on a new machine):

1. Downloads the pinned release (`v0.1.2`) for this Mac
   (`volatile-downloads-darwin-arm64` or `-amd64`).
2. Installs it to `~/.local/bin/volatile-downloads` **and** `~/go/bin/volatile-downloads`.
3. Loads LaunchAgent `com.zbornheimer.volatile-downloads` (runs at login).

That lives in the [dotfiles](https://github.com/zachbornheimer/dotfiles) repo:

- `home/run_onchange_after_install-volatile-downloads.sh.tmpl`
- `home/Library/LaunchAgents/com.zbornheimer.volatile-downloads.plist.tmpl`

Bump the `# rev:` line and `pin=` in the install script when cutting a new
release, then `chezmoi apply`.

To build from a checkout instead of the release:

```bash
VOLATILE_DOWNLOADS_FROM_SOURCE=1 chezmoi apply
```

## What it does at login

1. Ensures `/tmp/Downloads` exists (mode 755, owned by you — no sudo).
2. Makes `~/Downloads` a symlink there. Already correct → no-op.
3. If `~/Downloads` is a real folder, moves its contents into `/tmp/Downloads`
   (incoming file wins on name collision), then replaces the folder with the
   symlink. It will not `rm -R` a real Downloads folder.
4. Stamps `/tmp/Downloads` with the system Downloads folder icon
   (`DownloadsFolder.icns`). Relaunches Dock only when the icon was missing or
   the symlink changed.

`/tmp` on macOS is not a ramdisk. Files usually survive reboot until the
system's periodic tmp cleanup. That is still the point: Downloads are not a
durable archive.

This replaced the 2014 Intel Automator applet (`Download-Workflow.app`) that
deleted `~/Downloads` every login, chmod 777, and SIGKILL'd Dock.

## Develop (optional)

```bash
git clone git@github.com:zachbornheimer/volatile-downloads.git
cd volatile-downloads
mise run test
mise run install          # local signed build → ~/.local/bin and ~/go/bin
```

```bash
go install github.com/zachbornheimer/volatile-downloads/cmd/volatile-downloads@v0.1.2
```

`go install` only puts the binary in `GOBIN`. It does not load the LaunchAgent.
Chezmoi does that.

## Flags

```
volatile-downloads --version
volatile-downloads --target /tmp/Downloads --link ~/Downloads
volatile-downloads --no-dock
```

`--no-dock` skips the Dock relaunch (tests).

Env: `VOLATILE_DOWNLOADS_TARGET`, `VOLATILE_DOWNLOADS_LINK`.
