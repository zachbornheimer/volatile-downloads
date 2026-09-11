// Command volatile-downloads points ~/Downloads at a tmp-backed directory
// so browser downloads stay disposable.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	evo "github.com/zachbornheimer/evident-output"

	"github.com/zachbornheimer/volatile-downloads/internal/ensure"
	"github.com/zachbornheimer/volatile-downloads/internal/files"
	"github.com/zachbornheimer/volatile-downloads/internal/look"
)

// Version is injected at build time via ldflags. Default is the
// development checkout.
var Version = "dev"

const (
	defaultTarget = "/tmp/Downloads"
	envTarget     = "VOLATILE_DOWNLOADS_TARGET"
	envLink       = "VOLATILE_DOWNLOADS_LINK"
)

func main() {
	cfg, showVersion, parseErr := parseFlags(os.Args[1:])
	evo.Init(evo.Config{
		Title:  "volatile-downloads",
		Stdout: os.Stderr,
	})
	evo.Main(func() error {
		return run(cfg, showVersion, parseErr)
	})
}

func parseFlags(args []string) (ensure.Config, bool, error) {
	fs := flag.NewFlagSet("volatile-downloads", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	showVersion := fs.Bool("version", false, "print version and exit")
	target := fs.String("target", "", "tmp-backed directory (default /tmp/Downloads)")
	link := fs.String("link", "", "Downloads path (default ~/Downloads)")
	noDock := fs.Bool("no-dock", false, "do not relaunch Dock after setting the folder icon")
	if err := fs.Parse(args); err != nil {
		return ensure.Config{}, false, err
	}
	cfg, err := resolveConfig(*target, *link, !*noDock)
	return cfg, *showVersion, err
}

func resolveConfig(targetFlag, linkFlag string, refreshDock bool) (ensure.Config, error) {
	target := firstNonEmpty(targetFlag, os.Getenv(envTarget), defaultTarget)
	link := firstNonEmpty(linkFlag, os.Getenv(envLink))
	if link == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ensure.Config{}, fmt.Errorf("home directory: %w", err)
		}
		link = filepath.Join(home, "Downloads")
	}
	return ensure.Config{Target: target, Link: link, RefreshDock: refreshDock}, nil
}

func run(cfg ensure.Config, showVersion bool, parseErr error) error {
	if parseErr == flag.ErrHelp {
		return nil
	}
	if parseErr != nil {
		evo.Task("parse flags").Block(parseErr.Error())
		return nil
	}
	if showVersion {
		fmt.Fprintf(os.Stdout, "volatile-downloads %s\n", Version)
		return nil
	}
	task := evo.Task("ensure downloads")
	plan, err := ensure.Run(files.OS{}, look.AppKit{}, cfg)
	if err != nil {
		return task.Failf("%w", err)
	}
	task.Done(plan.Reason)
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
