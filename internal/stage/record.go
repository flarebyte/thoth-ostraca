package stage

// Record is the standard per-record shape in the envelope.
// Using a struct ensures deterministic JSON field ordering.
type Record struct {
	Locator  string         `json:"locator"`
	Meta     map[string]any `json:"meta,omitempty"`
	Mapped   any            `json:"mapped,omitempty"`
	Shell    *ShellResult   `json:"shell,omitempty"`
	Post     any            `json:"post,omitempty"`
	FileInfo *RecFileInfo   `json:"fileInfo,omitempty"`
	Git      *RecGit        `json:"git,omitempty"`
	Error    *RecError      `json:"error,omitempty"`
}

// ShellResult captures deterministic outputs of a shell execution.
type ShellResult struct {
	ExitCode int    `json:"exitCode"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

// RecFileInfo holds basic file metadata for a locator.
type RecFileInfo struct {
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	IsDir   bool   `json:"isDir"`
}

// RecGit holds minimal git metadata for a locator.
type RecGit struct {
	Tracked    bool          `json:"tracked"`
	Ignored    bool          `json:"ignored"`
	Status     string        `json:"status"`
	LastCommit *RecGitCommit `json:"lastCommit"`
}

type RecGitCommit struct {
	Hash   string `json:"hash"`
	Author string `json:"author"`
	Time   string `json:"time"`
}
