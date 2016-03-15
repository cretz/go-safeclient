package client

import (
	"strconv"
	"net/url"
	"errors"
)

type CreateDirInfo struct {
	DirPath string `json:"dirPath"`
	Private bool `json:"isPrivate"`
	Versioned bool `json:"isVersioned"`
	Metadata string `json:"metadata"`
	Shared bool `json:"isPathShared"`
}

func (c *Client) CreateDir(cd CreateDirInfo) error {
	req := &ClientRequest{
		Path: "/nfs/directory",
		Method: "POST",
		JSONBody: cd,
	}
	_, err := c.Do(req)
	return err
}

type GetDirInfo struct {
	DirPath string
	Shared bool
}

func (c *Client) GetDir(gd GetDirInfo) (DirResponse, error) {
	var resp DirResponse
	req := &ClientRequest{
		Path: "/nfs/directory/" + url.QueryEscape(gd.DirPath) + "/" + strconv.FormatBool(gd.Shared),
		Method: "GET",
		JSONResponse: &resp,
	}
	_, err := c.Do(req)
	return resp, err
}

type DeleteDirInfo struct {
	DirPath string
	Shared bool
}

func (c *Client) DeleteDir(dd DeleteDirInfo) error {
	req := &ClientRequest{
		Path: "/nfs/directory/" + url.QueryEscape(dd.DirPath) + "/" + strconv.FormatBool(dd.Shared),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}

type ChangeDirInfo struct {
	DirPath string `json:"-"`
	Shared bool `json:"-"`
	NewName string `json:"name,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// Empty NewName and Metadata are ignored
func (c *Client) ChangeDir(cd ChangeDirInfo) error {
	if cd.NewName == "" && cd.Metadata == "" {
		return errors.New("Must provide name or metadata")
	}
	req := &ClientRequest{
		Path: "/nfs/directory/" + url.QueryEscape(cd.DirPath) + "/" + strconv.FormatBool(cd.Shared),
		Method: "PUT",
		JSONBody: cd,
	}
	_, err := c.Do(req)
	return err
}

type MoveDirInfo struct {
	SrcPath string `json:"srcPath"`
	SrcShared bool `json:"isSrcPathShared"`
	DestPath string `json:"destPath"`
	DestShared bool `json:"isDestPathShared"`
	RetainSource bool `json:"retainSource"`
}

func (c *Client) MoveDir(md MoveDirInfo) error {
	// TODO: appears broken
	req := &ClientRequest{
		Path: "/nfs/movedir",
		Method: "POST",
		JSONBody: md,
	}
	_, err := c.Do(req)
	return err
}