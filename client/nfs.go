package client

import "time"

// Time is a representation of SAFE time (milliseconds since epoch)
type Time int64

// Time returns the Go time from the SAFE time
func (t Time) Time() time.Time {
	return time.Unix(0, int64(t)*int64(time.Millisecond))
}

// Files is a collection of files that implements sort.Interface
type Files []FileInfo

// Dirs is a collection of directories that implements sort.Interface
type Dirs []DirInfo

// DirResponse is a response from directory commands for a Client
type DirResponse struct {
	// Info about the directory requested
	Info DirInfo `json:"info"`
	// All files inside the directory
	Files Files `json:"files"`
	// All sub directories in the directory
	SubDirs Dirs `json:"subDirectories"`
}

// DirInfo is information about a single directory
type DirInfo struct {
	Name       string `json:"name"`
	Private    bool   `json:"isPrivate"`
	Versioned  bool   `json:"isVersioned"`
	CreatedOn  Time   `json:"createdOn"`
	ModifiedOn Time   `json:"modifiedOn"`
	Metadata   string `json:"metadata"`
}

// FileInfo is information about a single directory
type FileInfo struct {
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	CreatedOn  Time   `json:"createdOn"`
	ModifiedOn Time   `json:"modifiedOn"`
	Metadata   string `json:"metadata"`
}

func (f Files) Len() int           { return len(f) }
func (f Files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Files) Less(i, j int) bool { return f[i].Name < f[j].Name }

func (d Dirs) Len() int           { return len(d) }
func (d Dirs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d Dirs) Less(i, j int) bool { return d[i].Name < d[j].Name }
