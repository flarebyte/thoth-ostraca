// File Guide for dev/ai agents:
// Purpose: Define the deterministic per-record payload that stages read, enrich, and serialize.
// Responsibilities:
// - Declare the standard record shape used across all stage pipelines.
// - Hold shell, fileInfo, git, post, and embedded error substructures.
// - Keep JSON field ordering stable by using structs instead of ad hoc maps.
// Architecture notes:
// - This is the shared record contract; adding fields here affects every stage and many goldens.
// - `ShellResult` includes diagnostic fields like program, workingDir, and args on purpose for debugging failed shell runs.
// - Optional enrichments are pointers so absent data stays omitted in JSON output.
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
	ExitCode        int      `json:"exitCode"`
	JSON            any      `json:"json,omitempty"`
	Stdout          *string  `json:"stdout,omitempty"`
	Stderr          *string  `json:"stderr,omitempty"`
	StdoutTruncated bool     `json:"stdoutTruncated"`
	StderrTruncated bool     `json:"stderrTruncated"`
	TimedOut        bool     `json:"timedOut"`
	Error           *string  `json:"error,omitempty"`
	Program         string   `json:"program,omitempty"`
	WorkingDir      string   `json:"workingDir,omitempty"`
	Args            []string `json:"args,omitempty"`
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
