package ensure_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zachbornheimer/volatile-downloads/internal/ensure"
	"github.com/zachbornheimer/volatile-downloads/internal/files"
)

func sandbox(t *testing.T) ensure.Config {
	t.Helper()
	root := t.TempDir()
	return ensure.Config{
		Target: filepath.Join(root, "tmp-downloads"),
		Link:   filepath.Join(root, "Downloads"),
	}
}

func TestRun_CreatesSymlinkWhenMissing(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionCreateLink {
		t.Fatalf("action = %v, want create", plan.Action)
	}
	assertDesiredSymlink(t, cfg)
}

func TestRun_NoopWhenAlreadyCorrect(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.MkdirAll(cfg.Target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cfg.Target, cfg.Link); err != nil {
		t.Fatal(err)
	}
	before, err := os.Lstat(cfg.Link)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionNothing {
		t.Fatalf("action = %v, want nothing", plan.Action)
	}
	after, err := os.Lstat(cfg.Link)
	if err != nil {
		t.Fatal(err)
	}
	if !before.ModTime().Equal(after.ModTime()) {
		t.Fatal("symlink was rewritten on no-op")
	}
}

func TestRun_ReplacesWrongSymlink(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	other := filepath.Join(filepath.Dir(cfg.Link), "other")
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(other, cfg.Link); err != nil {
		t.Fatal(err)
	}
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionReplaceSymlink {
		t.Fatalf("action = %v, want replace symlink", plan.Action)
	}
	assertDesiredSymlink(t, cfg)
}

func TestRun_ReplacesEmptyDirectory(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.MkdirAll(cfg.Link, 0o755); err != nil {
		t.Fatal(err)
	}
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionMergeAndReplace {
		t.Fatalf("action = %v, want merge/replace", plan.Action)
	}
	assertDesiredSymlink(t, cfg)
}

func TestRun_MergesExistingFiles(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.MkdirAll(cfg.Link, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Link, "receipt.pdf"), []byte("kept"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionMergeAndReplace {
		t.Fatalf("action = %v, want merge/replace", plan.Action)
	}
	assertDesiredSymlink(t, cfg)
	got, err := os.ReadFile(filepath.Join(cfg.Target, "receipt.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "kept" {
		t.Fatalf("merged content = %q", got)
	}
}

func TestRun_NameCollisionKeepsIncomingDownloadsFile(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.MkdirAll(cfg.Link, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cfg.Target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Target, "receipt.pdf"), []byte("stale-tmp"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Link, "receipt.pdf"), []byte("from-downloads"), 0o644); err != nil {
		t.Fatal(err)
	}
	plan, err := ensure.Run(files.OS{}, nil, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionMergeAndReplace {
		t.Fatalf("action = %v, want merge/replace", plan.Action)
	}
	got, err := os.ReadFile(filepath.Join(cfg.Target, "receipt.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "from-downloads" {
		t.Fatalf("merged content = %q, want incoming Downloads file", got)
	}
	assertDesiredSymlink(t, cfg)
}

func TestRun_RemoveAllNeverTargetsDownloadsPath(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.MkdirAll(cfg.Link, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.Link, "a.txt"), []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec := &recordingFS{FS: files.OS{}}
	if _, err := ensure.Run(rec, nil, cfg); err != nil {
		t.Fatalf("Run: %v", err)
	}
	for _, path := range rec.removeAll {
		if path == cfg.Link {
			t.Fatalf("RemoveAll(%q) targeted the Downloads path", path)
		}
	}
}

type recordingFS struct {
	ensure.FS
	removeAll []string
}

func (r *recordingFS) RemoveAll(path string) error {
	r.removeAll = append(r.removeAll, path)
	return r.FS.RemoveAll(path)
}

func TestRun_RefusesRegularFile(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.WriteFile(cfg.Link, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := ensure.Run(files.OS{}, nil, cfg)
	if err == nil {
		t.Fatal("expected refuse error")
	}
	info, err := os.Lstat(cfg.Link)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode()&os.ModeSymlink != 0 || info.IsDir() {
		t.Fatal("regular file was replaced")
	}
}

func TestRun_AppliesDownloadsIconOnNoop(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	cfg.RefreshDock = true
	if err := os.MkdirAll(cfg.Target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(cfg.Target, cfg.Link); err != nil {
		t.Fatal(err)
	}
	look := &recordLook{}
	plan, err := ensure.Run(files.OS{}, look, cfg)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if plan.Action != ensure.ActionNothing {
		t.Fatalf("action = %v, want nothing", plan.Action)
	}
	if len(look.dirs) != 1 || look.dirs[0] != cfg.Target {
		t.Fatalf("icon dirs = %v, want [%s]", look.dirs, cfg.Target)
	}
	if look.docks != 1 {
		t.Fatalf("dock refreshes = %d, want 1 on first icon apply", look.docks)
	}
}

func TestRun_SkipsLookWhenRefused(t *testing.T) {
	t.Parallel()
	cfg := sandbox(t)
	if err := os.WriteFile(cfg.Link, []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	look := &recordLook{}
	if _, err := ensure.Run(files.OS{}, look, cfg); err == nil {
		t.Fatal("expected refuse error")
	}
	if len(look.dirs) != 0 || look.docks != 0 {
		t.Fatalf("look ran on refuse: dirs=%v docks=%d", look.dirs, look.docks)
	}
}

type recordLook struct {
	dirs  []string
	docks int
}

func (r *recordLook) ApplyDownloadsIcon(dir string) error {
	r.dirs = append(r.dirs, dir)
	return nil
}

func (r *recordLook) RefreshDock() error {
	r.docks++
	return nil
}

func assertDesiredSymlink(t *testing.T, cfg ensure.Config) {
	t.Helper()
	dest, err := os.Readlink(cfg.Link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if dest != cfg.Target {
		t.Fatalf("link dest = %q, want %q", dest, cfg.Target)
	}
	if _, err := os.Stat(cfg.Target); err != nil {
		t.Fatalf("target missing: %v", err)
	}
}
