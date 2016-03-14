package client

import "time"

type Time int64

func (t Time) Time() time.Time {
	return time.Unix(0, int64(t) * int64(time.Millisecond))
}

type DirResponse struct {
	Info DirInfo `json:"info"`
	Files []FileInfo `json:"files"`
	SubDirs []DirInfo `json:"subDirectories"`
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
