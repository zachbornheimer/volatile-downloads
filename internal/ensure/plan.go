package ensure

// LinkKind is what currently occupies the Downloads path.
type LinkKind int

const (
	// LinkMissing means the path does not exist.
	LinkMissing LinkKind = iota
	// LinkSymlink means the path is a symlink.
	LinkSymlink
	// LinkDir means the path is a real directory.
	LinkDir
	// LinkOther means the path is a regular file or something we will not replace.
	LinkOther
)

// Observation is the live state of the target directory and the Downloads path.
type Observation struct {
	TargetExists  bool
	TargetHasIcon bool
	Link          LinkKind
	LinkDest      string
	DirNames      []string
}

// Action is the effect Execute should perform.
type Action int

const (
	// ActionNothing: target exists and the symlink already points at it.
	ActionNothing Action = iota
	// ActionCreateLink: path is missing; create the symlink.
	ActionCreateLink
	// ActionReplaceSymlink: path is a symlink to the wrong place.
	ActionReplaceSymlink
	// ActionMergeAndReplace: path is a real directory; move contents, then symlink.
	ActionMergeAndReplace
	// ActionRefuse: path is a file or other node we will not touch.
	ActionRefuse
)

// Plan is the pure decision produced from an Observation.
type Plan struct {
	Action      Action
	Reason      string
	Merge       []string
	RefreshDock bool
}

const (
	dsStoreName  = ".DS_Store"
	iconFileName = "Icon\r"
)

func desiredLink(dest, target string) bool {
	return dest == target || dest == target+"/"
}

func skipName(name string) bool {
	return name == dsStoreName || name == iconFileName
}

func mergeNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if skipName(name) {
			continue
		}
		out = append(out, name)
	}
	return out
}

// Decide returns the plan for obs. It performs no I/O.
func Decide(obs Observation, cfg Config) Plan {
	p := Plan{}
	switch obs.Link {
	case LinkOther:
		p.Action = ActionRefuse
		p.Reason = "Downloads exists and is not a directory or symlink"
	case LinkSymlink:
		if desiredLink(obs.LinkDest, cfg.Target) {
			p.Action = ActionNothing
			p.Reason = "Downloads already points at " + cfg.Target
		} else {
			p.Action = ActionReplaceSymlink
			p.Reason = "replace symlink so Downloads points at " + cfg.Target
		}
	case LinkDir:
		p.Action = ActionMergeAndReplace
		p.Merge = mergeNames(obs.DirNames)
		if len(p.Merge) == 0 {
			p.Reason = "replace empty Downloads directory with symlink"
		} else {
			p.Reason = "merge existing Downloads into " + cfg.Target
		}
	default:
		p.Action = ActionCreateLink
		p.Reason = "point Downloads at " + cfg.Target
	}
	p.RefreshDock = cfg.RefreshDock && p.Action != ActionRefuse &&
		(!obs.TargetHasIcon || p.Action != ActionNothing)
	return p
}
