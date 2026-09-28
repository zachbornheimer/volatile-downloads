package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const ensureExitDeadline = 5 * time.Second

func TestVersionFlagPrintsInjectedVersion(t *testing.T) {
	t.Parallel()
	bin := buildBin(t)
	out, err := exec.Command(bin, "--version").Output()
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	got := strings.TrimSpace(string(out))
	if got != "volatile-downloads test-version" {
		t.Fatalf("version = %q", got)
	}
}

func TestBinaryCreatesSymlinkInSandbox(t *testing.T) {
	t.Parallel()
	bin := buildBin(t)
	root := t.TempDir()
	target := filepath.Join(root, "tmp-downloads")
	link := filepath.Join(root, "Downloads")
	cmd := exec.Command(bin, "--no-dock", "--target", target, "--link", link)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	dest, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if dest != target {
		t.Fatalf("link dest = %q, want %q", dest, target)
	}
}

func TestBinaryExitsAfterEnsure(t *testing.T) {
	t.Parallel()
	bin := buildBin(t)
	root := t.TempDir()
	target := filepath.Join(root, "tmp-downloads")
	link := filepath.Join(root, "Downloads")
	ctx, cancel := context.WithTimeout(t.Context(), ensureExitDeadline)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--no-dock", "--target", target, "--link", link)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			t.Fatalf("binary did not exit within %s: %v\n%s", ensureExitDeadline, err, out)
		}
		t.Fatalf("run: %v\n%s", err, out)
	}
	dest, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if dest != target {
		t.Fatalf("link dest = %q, want %q", dest, target)
	}
}

func buildBin(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "volatile-downloads")
	cmd := exec.Command("go", "build", "-ldflags", "-X main.Version=test-version", "-o", bin, ".")
	cmd.Dir = filepath.Join(".", "")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}
