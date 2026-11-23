package look_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/zachbornheimer/volatile-downloads/internal/look"
)

func TestAppKit_ApplyDownloadsIconWritesIconFile(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("AppKit is macOS-only")
	}
	dir := t.TempDir()
	if err := (look.AppKit{}).ApplyDownloadsIcon(dir); err != nil {
		t.Fatalf("ApplyDownloadsIcon: %v", err)
	}
	info, err := os.Lstat(filepath.Join(dir, "Icon\r"))
	if err != nil {
		t.Fatalf("Icon file missing: %v", err)
	}
	if info.IsDir() {
		t.Fatal("Icon file is a directory")
	}
}
