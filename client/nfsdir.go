package client

import (
	"errors"
	"net/url"
	"strconv"
)

// CreateDirInfo are parameters for Client.CreateDir
type CreateDirInfo struct {
	// The path to create
	DirPath string `json:"dirPath"`
	// Whether the directory is private
	Private bool `json:"isPrivate"`
	// Whether the directory is versioned
	Versioned bool `json:"isVersioned"`
	// Metadata to store with the directory
	Metadata string `json:"metadata"`
	// Whether the directory is shared
	Shared bool `json:"isPathShared"`
}

// CreateDir creates a directory. See https://maidsafe.readme.io/docs/nfs-create-directory for more info.
func (c *Client) CreateDir(cd CreateDirInfo) error {
	req := &Request{
		Path:     "/nfs/directory",
		Method:   "POST",
		JSONBody: cd,
	}
	_, err := c.Do(req)
	return err
}

// GetDirInfo are parameters for Client.GetDir
type GetDirInfo struct {
	// The path to fetch
	DirPath string
	// Whether the directory is shared
	Shared bool
}

// GetDir gets directory information. See https://maidsafe.readme.io/docs/nfs-get-directory for more info.
func (c *Client) GetDir(gd GetDirInfo) (DirResponse, error) {
	var resp DirResponse
	req := &Request{
		Path:         "/nfs/directory/" + url.QueryEscape(gd.DirPath) + "/" + strconv.FormatBool(gd.Shared),
		Method:       "GET",
		JSONResponse: &resp,
	}
	_, err := c.Do(req)
	return resp, err
}

// DeleteDirInfo are parameters for Client.DeleteDir
type DeleteDirInfo struct {
	// The path to delete
	DirPath string
	// Whether the path is shared
	Shared bool
}

// DeleteDir deletes a directory. See https://maidsafe.readme.io/docs/nfs-delete-directory for more info.
func (c *Client) DeleteDir(dd DeleteDirInfo) error {
	req := &Request{
		Path:   "/nfs/directory/" + url.QueryEscape(dd.DirPath) + "/" + strconv.FormatBool(dd.Shared),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}

// ChangeDirInfo are parameters to ChangeDir. One of NewName or Metadata must be present.
type ChangeDirInfo struct {
	// The path to change
	DirPath string `json:"-"`
	// Whether the directory is shared
	Shared bool `json:"-"`
	// The name to change it to. If Metadata is not set, this must be set.
	NewName string `json:"name,omitempty"`
	// The metadata to set. If NewName is not set, this must be set.
	Metadata string `json:"metadata,omitempty"`
}

// ChangeDir changes a directory's name, metadata, or both. There is no documentation for this. See
// https://maidsafe.atlassian.net/browse/CS-60 for more information.
func (c *Client) ChangeDir(cd ChangeDirInfo) error {
	// TODO: how to change only the name and not the metadata?
	if cd.NewName == "" && cd.Metadata == "" {
		return errors.New("Must provide name or metadata")
	}
	req := &Request{
		Path:     "/nfs/directory/" + url.QueryEscape(cd.DirPath) + "/" + strconv.FormatBool(cd.Shared),
		Method:   "PUT",
		JSONBody: cd,
	}
	_, err := c.Do(req)
	return err
}

// MoveDirInfo are parameters to MoveDir. This is currently undocumented/unsupported. See
// https://maidsafe.atlassian.net/browse/CS-60 for more info.
type MoveDirInfo struct {
	SrcPath      string `json:"srcPath"`
	SrcShared    bool   `json:"isSrcPathShared"`
	DestPath     string `json:"destPath"`
	DestShared   bool   `json:"isDestPathShared"`
	RetainSource bool   `json:"retainSource"`
}

// MoveDir moves a directory. This is currently undocumented/unsupported. See
// https://maidsafe.atlassian.net/browse/CS-60 for more info.
func (c *Client) MoveDir(md MoveDirInfo) error {
	req := &Request{
		Path:     "/nfs/movedir",
		Method:   "POST",
		JSONBody: md,
	}
	_, err := c.Do(req)
	return err
}
