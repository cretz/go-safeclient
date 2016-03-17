package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
)

// CreateFileInfo are parameters for Client.CreateFile
type CreateFileInfo struct {
	// The path to create
	FilePath string `json:"filePath"`
	// Whether the path is shared
	Shared bool `json:"isPathShared"`
	// Metadata to store with the file
	Metadata string `json:"metadata"`
}

// CreateFile creates a file. See https://maidsafe.readme.io/docs/nfsfile for more information.
func (c *Client) CreateFile(cf CreateFileInfo) error {
	req := &Request{
		Path:     "/nfs/file",
		Method:   "POST",
		JSONBody: cf,
	}
	_, err := c.Do(req)
	return err
}

// MoveFileInfo are parameters for Client.MoveFile. This is currently undocumented/unsupported. See
// https://maidsafe.atlassian.net/browse/CS-60 for more info.
type MoveFileInfo struct {
	SrcPath      string `json:"srcPath"`
	SrcShared    bool   `json:"isSrcPathShared"`
	DestPath     string `json:"destPath"`
	DestShared   bool   `json:"isDestPathShared"`
	RetainSource bool   `json:"retainSource"`
}

// MoveFile moves a file. This is currently undocumented/unsupported. See https://maidsafe.atlassian.net/browse/CS-60
// for more info.
func (c *Client) MoveFile(mf MoveFileInfo) error {
	// TODO: appears broken
	req := &Request{
		Path:     "/nfs/movefile",
		Method:   "POST",
		JSONBody: mf,
	}
	_, err := c.Do(req)
	return err
}

// DeleteFileInfo are parameters for Client.DeleteFile
type DeleteFileInfo struct {
	// The path to delete
	FilePath string
	// Whether the path is shared
	Shared bool
}

// DeleteFile deletes a file. See https://maidsafe.readme.io/docs/nfs-delete-file for more info.
func (c *Client) DeleteFile(df DeleteFileInfo) error {
	req := &Request{
		Path:   "/nfs/file/" + url.QueryEscape(df.FilePath) + "/" + strconv.FormatBool(df.Shared),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}

// ChangeFileInfo are parameters for Client.ChangeFile. This is currently undocumented/unsupported. See
// https://maidsafe.atlassian.net/browse/CS-60 for more info.
type ChangeFileInfo struct {
	FilePath string `json:"-"`
	Shared   bool   `json:"-"`
	NewName  string `json:"name,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// ChangeFile changes the file name, metadata, or both. This is currently undocumented/unsupported. See
// https://maidsafe.atlassian.net/browse/CS-60 for more info.
func (c *Client) ChangeFile(cf ChangeFileInfo) error {
	// TODO: what if they want to set metadata to empty?
	if cf.NewName == "" && cf.Metadata == "" {
		return errors.New("Must provide name or metadata")
	}
	req := &Request{
		Path:     "/nfs/file/metadata/" + url.QueryEscape(cf.FilePath) + "/" + strconv.FormatBool(cf.Shared),
		Method:   "PUT",
		JSONBody: cf,
	}
	_, err := c.Do(req)
	return err
}

// WriteFileInfo are parameters for Client.WriteFile
type WriteFileInfo struct {
	// The path to write to
	FilePath string
	// Whether the path is shared
	Shared bool
	// The contents to write
	Contents io.ReadCloser
	// The byte offset in the file to start writing
	Offset int64
}

// WriteFile writes a file. See https://maidsafe.readme.io/docs/nfs-update-file-content for more info.
func (c *Client) WriteFile(wf WriteFileInfo) error {
	// TODO: support chunking instead of all in mem
	defer wf.Contents.Close()
	byts, err := ioutil.ReadAll(wf.Contents)
	if err != nil {
		return fmt.Errorf("Unable to read contents: %v", err)
	}
	req := &Request{
		Path:    "/nfs/file/" + url.QueryEscape(wf.FilePath) + "/" + strconv.FormatBool(wf.Shared),
		Method:  "PUT",
		RawBody: byts,
		Query:   map[string][]string{"offset": []string{strconv.FormatInt(wf.Offset, 10)}},
	}
	_, err = c.Do(req)
	return err
}

// GetFileInfo are parameters for Client.GetFile
type GetFileInfo struct {
	// The path to get
	FilePath string
	// Whether the path is shared
	Shared bool
	// The byte offset to start reading at
	Offset int64
	// The amount of bytes to read. If the value is 0 then there is no length constraint
	Length int64
}

// GetFile obtains a file's contents. See https://maidsafe.readme.io/docs/nfs-get-file for more info.
func (c *Client) GetFile(gf GetFileInfo) (io.ReadCloser, error) {
	// TODO: support chunking instead of all in mem
	query := map[string][]string{"offset": []string{strconv.FormatInt(gf.Offset, 10)}}
	if gf.Length > 0 {
		query["length"] = []string{strconv.FormatInt(gf.Length, 10)}
	}
	req := &Request{
		Path:   "/nfs/file/" + url.QueryEscape(gf.FilePath) + "/" + strconv.FormatBool(gf.Shared),
		Method: "GET",
		Query:  query,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Unable to get file: %v", err)
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Unable to read file output: %v", err)
	}
	return ioutil.NopCloser(bytes.NewReader(byts)), nil
}
