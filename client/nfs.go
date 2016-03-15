package client

import "time"

type Time int64

func (t Time) Time() time.Time {
	return time.Unix(0, int64(t) * int64(time.Millisecond))
}

type Files []FileInfo
type Dirs []DirInfo

type DirResponse struct {
	Info DirInfo `json:"info"`
	Files Files `json:"files"`
	SubDirs Dirs `json:"subDirectories"`
}

type DirInfo struct {
	Name string `json:"name"`
	Private bool `json:"isPrivate"`
	Versioned bool `json:"isVersioned"`
	CreatedOn Time `json:"createdOn"`
	ModifiedOn Time `json:"modifiedOn"`
	Metadata string `json:"metadata"`
}

type FileInfo struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	CreatedOn Time `json:"createdOn"`
	ModifiedOn Time `json:"modifiedOn"`
	Metadata string `json:"metadata"`
}

func (f Files) Len() int           { return len(f) }
func (f Files) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f Files) Less(i, j int) bool { return f[i].Name < f[j].Name }

func (d Dirs) Len() int           { return len(d) }
func (d Dirs) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d Dirs) Less(i, j int) bool { return d[i].Name < d[j].Name }