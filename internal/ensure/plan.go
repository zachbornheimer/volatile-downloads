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
	TargetExists bool
	Link         LinkKind
	LinkDest     string
	DirNames     []string
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
	Action Action
	Reason string
	Merge  []string
}

const dsStoreName = ".DS_Store"

func desiredLink(dest, target string) bool {
	return dest == target || dest == target+"/"
}

func mergeNames(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if name == dsStoreName {
			continue
		}
		out = append(out, name)
	}
	return out
}

// Decide returns the plan for obs. It performs no I/O.
func Decide(obs Observation, cfg Config) Plan {
	switch obs.Link {
	case LinkOther:
		return Plan{Action: ActionRefuse, Reason: "Downloads exists and is not a directory or symlink"}
	case LinkSymlink:
		if desiredLink(obs.LinkDest, cfg.Target) {
			return Plan{Action: ActionNothing, Reason: "Downloads already points at " + cfg.Target}
		}
		return Plan{Action: ActionReplaceSymlink, Reason: "replace symlink so Downloads points at " + cfg.Target}
	case LinkDir:
		names := mergeNames(obs.DirNames)
		if len(names) == 0 {
			return Plan{Action: ActionMergeAndReplace, Reason: "replace empty Downloads directory with symlink"}
		}
		return Plan{
			Action: ActionMergeAndReplace,
			Reason: "merge existing Downloads into " + cfg.Target,
			Merge:  names,
		}
	default:
		return Plan{Action: ActionCreateLink, Reason: "point Downloads at " + cfg.Target}
	}
}
