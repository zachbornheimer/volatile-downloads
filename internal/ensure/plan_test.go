package ensure

import "testing"

func TestDecide_AlreadyCorrectSymlink(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{TargetExists: true, Link: LinkSymlink, LinkDest: "/tmp/Downloads"}, cfg)
	if plan.Action != ActionNothing {
		t.Fatalf("action = %v, want nothing", plan.Action)
	}
}

func TestDecide_TrailingSlashOnTargetStillMatches(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkSymlink, LinkDest: "/tmp/Downloads/"}, cfg)
	if plan.Action != ActionNothing {
		t.Fatalf("action = %v, want nothing", plan.Action)
	}
}

func TestDecide_WrongSymlinkIsReplaced(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkSymlink, LinkDest: "/var/tmp/other"}, cfg)
	if plan.Action != ActionReplaceSymlink {
		t.Fatalf("action = %v, want replace symlink", plan.Action)
	}
}

func TestDecide_MissingPathCreatesLink(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkMissing}, cfg)
	if plan.Action != ActionCreateLink {
		t.Fatalf("action = %v, want create", plan.Action)
	}
}

func TestDecide_EmptyDirIsReplaced(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkDir, DirNames: []string{".DS_Store"}}, cfg)
	if plan.Action != ActionMergeAndReplace {
		t.Fatalf("action = %v, want merge/replace", plan.Action)
	}
	if len(plan.Merge) != 0 {
		t.Fatalf("merge = %v, want empty (.DS_Store ignored)", plan.Merge)
	}
}

func TestDecide_DirWithFilesMerges(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkDir, DirNames: []string{"receipt.pdf", ".DS_Store"}}, cfg)
	if plan.Action != ActionMergeAndReplace {
		t.Fatalf("action = %v, want merge/replace", plan.Action)
	}
	if len(plan.Merge) != 1 || plan.Merge[0] != "receipt.pdf" {
		t.Fatalf("merge = %v, want [receipt.pdf]", plan.Merge)
	}
}

func TestDecide_RegularFileIsRefused(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkOther}, cfg)
	if plan.Action != ActionRefuse {
		t.Fatalf("action = %v, want refuse", plan.Action)
	}
}

func TestDecide_SkipsIconFileWhenMerging(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads"}
	plan := Decide(Observation{Link: LinkDir, DirNames: []string{"receipt.pdf", "Icon\r", ".DS_Store"}}, cfg)
	if len(plan.Merge) != 1 || plan.Merge[0] != "receipt.pdf" {
		t.Fatalf("merge = %v, want [receipt.pdf]", plan.Merge)
	}
}

func TestDecide_RefreshDockWhenIconMissing(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads", RefreshDock: true}
	plan := Decide(Observation{TargetExists: true, TargetHasIcon: false, Link: LinkSymlink, LinkDest: "/tmp/Downloads"}, cfg)
	if !plan.RefreshDock {
		t.Fatal("expected Dock refresh when the Downloads icon is missing")
	}
}

func TestDecide_NoDockRefreshWhenIconAlreadyPresent(t *testing.T) {
	t.Parallel()
	cfg := Config{Target: "/tmp/Downloads", Link: "/Users/z/Downloads", RefreshDock: true}
	plan := Decide(Observation{TargetExists: true, TargetHasIcon: true, Link: LinkSymlink, LinkDest: "/tmp/Downloads"}, cfg)
	if plan.RefreshDock {
		t.Fatal("did not expect Dock refresh when the icon is already set")
	}
}
