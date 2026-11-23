package ensure

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// TargetPerm is the mode for the tmp-backed Downloads directory.
const TargetPerm fs.FileMode = 0o755

// Config names the tmp directory and the Downloads path that should point at it.
type Config struct {
	Target      string
	Link        string
	RefreshDock bool
}

// Look applies the system Downloads folder icon and refreshes the Dock.
type Look interface {
	ApplyDownloadsIcon(dir string) error
	RefreshDock() error
}

// FS is the filesystem port. ensure defines it; files.OS implements it.
type FS interface {
	Lstat(path string) (fs.FileInfo, error)
	Readlink(path string) (string, error)
	ReadDir(path string) ([]fs.DirEntry, error)
	MkdirAll(path string, perm fs.FileMode) error
	Chmod(path string, perm fs.FileMode) error
	Rename(src, dst string) error
	Remove(path string) error
	RemoveAll(path string) error
	Symlink(oldname, newname string) error
}

// Observe gathers the live state of cfg.Target and cfg.Link.
func Observe(fsys FS, cfg Config) (Observation, error) {
	var obs Observation
	if _, err := fsys.Lstat(cfg.Target); err == nil {
		obs.TargetExists = true
		if _, err := fsys.Lstat(filepath.Join(cfg.Target, iconFileName)); err == nil {
			obs.TargetHasIcon = true
		}
	} else if !os.IsNotExist(err) {
		return Observation{}, fmt.Errorf("stat target %q: %w", cfg.Target, err)
	}

	info, err := fsys.Lstat(cfg.Link)
	switch {
	case os.IsNotExist(err):
		obs.Link = LinkMissing
		return obs, nil
	case err != nil:
		return Observation{}, fmt.Errorf("stat link %q: %w", cfg.Link, err)
	}

	mode := info.Mode()
	switch {
	case mode&fs.ModeSymlink != 0:
		obs.Link = LinkSymlink
		dest, err := fsys.Readlink(cfg.Link)
		if err != nil {
			return Observation{}, fmt.Errorf("readlink %q: %w", cfg.Link, err)
		}
		obs.LinkDest = dest
	case info.IsDir():
		obs.Link = LinkDir
		entries, err := fsys.ReadDir(cfg.Link)
		if err != nil {
			return Observation{}, fmt.Errorf("readdir %q: %w", cfg.Link, err)
		}
		obs.DirNames = make([]string, 0, len(entries))
		for _, e := range entries {
			obs.DirNames = append(obs.DirNames, e.Name())
		}
	default:
		obs.Link = LinkOther
	}
	return obs, nil
}

// Execute applies plan. It does not invent policy.
func Execute(fsys FS, look Look, cfg Config, plan Plan) error {
	if plan.Action == ActionRefuse {
		return fmt.Errorf("%s", plan.Reason)
	}
	if err := ensureTarget(fsys, cfg.Target); err != nil {
		return err
	}
	switch plan.Action {
	case ActionNothing:
	case ActionCreateLink:
		if err := createLink(fsys, cfg); err != nil {
			return err
		}
	case ActionReplaceSymlink:
		if err := fsys.Remove(cfg.Link); err != nil {
			return fmt.Errorf("remove symlink %q: %w", cfg.Link, err)
		}
		if err := createLink(fsys, cfg); err != nil {
			return err
		}
	case ActionMergeAndReplace:
		if err := mergeDir(fsys, cfg, plan.Merge); err != nil {
			return err
		}
		if err := createLink(fsys, cfg); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown action %d", plan.Action)
	}
	return applyLook(look, cfg.Target, plan.RefreshDock)
}

// Run observes, decides, and executes.
func Run(fsys FS, look Look, cfg Config) (Plan, error) {
	obs, err := Observe(fsys, cfg)
	if err != nil {
		return Plan{}, err
	}
	plan := Decide(obs, cfg)
	if err := Execute(fsys, look, cfg, plan); err != nil {
		return plan, err
	}
	return plan, nil
}

func applyLook(look Look, dir string, refreshDock bool) error {
	if look == nil {
		return nil
	}
	if err := look.ApplyDownloadsIcon(dir); err != nil {
		return fmt.Errorf("downloads icon: %w", err)
	}
	if refreshDock {
		_ = look.RefreshDock()
	}
	return nil
}

func ensureTarget(fsys FS, target string) error {
	if err := fsys.MkdirAll(target, TargetPerm); err != nil {
		return fmt.Errorf("mkdir %q: %w", target, err)
	}
	if err := fsys.Chmod(target, TargetPerm); err != nil {
		return fmt.Errorf("chmod %q: %w", target, err)
	}
	return nil
}

func createLink(fsys FS, cfg Config) error {
	if err := fsys.Symlink(cfg.Target, cfg.Link); err != nil {
		return fmt.Errorf("symlink %q -> %q: %w", cfg.Link, cfg.Target, err)
	}
	return nil
}

func mergeDir(fsys FS, cfg Config, names []string) error {
	for _, name := range names {
		src := filepath.Join(cfg.Link, name)
		dst := filepath.Join(cfg.Target, name)
		if _, err := fsys.Lstat(dst); err == nil {
			if err := fsys.RemoveAll(dst); err != nil {
				return fmt.Errorf("replace leftover %q: %w", dst, err)
			}
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("stat %q: %w", dst, err)
		}
		if err := fsys.Rename(src, dst); err != nil {
			return fmt.Errorf("move %q to %q: %w", src, dst, err)
		}
	}
	_ = fsys.Remove(filepath.Join(cfg.Link, dsStoreName))
	_ = fsys.Remove(filepath.Join(cfg.Link, iconFileName))
	if err := fsys.Remove(cfg.Link); err != nil {
		return fmt.Errorf("replace directory %q: %w", cfg.Link, err)
	}
	return nil
}
