// Package look applies macOS folder appearance via AppKit.
package look

import (
	"fmt"
	"os/exec"
)

// DownloadsFolderIcon is the system Downloads folder icns.
const DownloadsFolderIcon = "/System/Library/CoreServices/CoreTypes.bundle/Contents/Resources/DownloadsFolder.icns"

const setIconJS = `function run(argv) {
  ObjC.import('AppKit');
  var dest = argv[0];
  var icns = argv[1];
  var img = $.NSImage.alloc.initWithContentsOfFile(icns);
  if (img.isNil()) {
    throw new Error('could not load Downloads folder icon');
  }
  var ok = $.NSWorkspace.sharedWorkspace.setIconForFileOptions(img, dest, 0);
  if (!ok) {
    throw new Error('NSWorkspace setIcon failed');
  }
}`

// AppKit sets folder icons through NSWorkspace and relaunches Dock.
type AppKit struct{}

// ApplyDownloadsIcon stamps dir with the built-in Downloads folder icon.
func (AppKit) ApplyDownloadsIcon(dir string) error {
	cmd := exec.Command("osascript", "-l", "JavaScript", "-e", setIconJS, "--", dir, DownloadsFolderIcon)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("set Downloads icon on %q: %w (%s)", dir, err, out)
	}
	return nil
}

// RefreshDock relaunches Dock so stacks pick up a new folder icon.
func (AppKit) RefreshDock() error {
	err := exec.Command("killall", "Dock").Run()
	if err != nil {
		return fmt.Errorf("relaunch Dock: %w", err)
	}
	return nil
}
