package client

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"strconv"
)

type CreateFileInfo struct {
	FilePath string `json:"filePath"`
	Shared   bool   `json:"isPathShared"`
	Metadata string `json:"metadata"`
}

func (c *Client) CreateFile(cf CreateFileInfo) error {
	req := &ClientRequest{
		Path:     "/nfs/file",
		Method:   "POST",
		JSONBody: cf,
	}
	_, err := c.Do(req)
	return err
}

type MoveFileInfo struct {
	SrcPath      string `json:"srcPath"`
	SrcShared    bool   `json:"isSrcPathShared"`
	DestPath     string `json:"destPath"`
	DestShared   bool   `json:"isDestPathShared"`
	RetainSource bool   `json:"retainSource"`
}

func (c *Client) MoveFile(mf MoveFileInfo) error {
	// TODO: appears broken
	req := &ClientRequest{
		Path:     "/nfs/movefile",
		Method:   "POST",
		JSONBody: mf,
	}
	_, err := c.Do(req)
	return err
}

type DeleteFileInfo struct {
	FilePath string
	Shared   bool
}

func (c *Client) DeleteFile(df DeleteFileInfo) error {
	req := &ClientRequest{
		Path:   "/nfs/file/" + url.QueryEscape(df.FilePath) + "/" + strconv.FormatBool(df.Shared),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}

type ChangeFileInfo struct {
	FilePath string `json:"-"`
	Shared   bool   `json:"-"`
	NewName  string `json:"name,omitempty"`
	Metadata string `json:"metadata,omitempty"`
}

// Empty NewName and Metadata are ignored
func (c *Client) ChangeFile(cf ChangeFileInfo) error {
	// TODO: what if they want to set metadata to empty?
	if cf.NewName == "" && cf.Metadata == "" {
		return errors.New("Must provide name or metadata")
	}
	req := &ClientRequest{
		Path:     "/nfs/file/metadata/" + url.QueryEscape(cf.FilePath) + "/" + strconv.FormatBool(cf.Shared),
		Method:   "PUT",
		JSONBody: cf,
	}
	_, err := c.Do(req)
	return err
}

type WriteFileInfo struct {
	FilePath string
	Shared   bool
	Contents io.ReadCloser
	Offset   int64
}

func (c *Client) WriteFile(wf WriteFileInfo) error {
	// TODO: support chunking instead of all in mem
	defer wf.Contents.Close()
	byts, err := ioutil.ReadAll(wf.Contents)
	if err != nil {
		return fmt.Errorf("Unable to read contents: %v", err)
	}
	req := &ClientRequest{
		Path:    "/nfs/file/" + url.QueryEscape(wf.FilePath) + "/" + strconv.FormatBool(wf.Shared),
		Method:  "PUT",
		RawBody: []byte(base64.StdEncoding.EncodeToString(byts)),
		Query:   map[string][]string{"offset": []string{strconv.FormatInt(wf.Offset, 10)}},
	}
	_, err = c.Do(req)
	return err
}

type GetFileInfo struct {
	FilePath string
	Shared   bool
	Offset   int64
	// Ignored if 0
	Length int64
}

func (c *Client) GetFile(gf GetFileInfo) (io.ReadCloser, error) {
	// TODO: support chunking instead of all in mem
	query := map[string][]string{"offset": []string{strconv.FormatInt(gf.Offset, 10)}}
	if gf.Length > 0 {
		query["length"] = []string{strconv.FormatInt(gf.Length, 10)}
	}
	req := &ClientRequest{
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
	byts, err = base64.StdEncoding.DecodeString(string(byts))
	if err != nil {
		return nil, fmt.Errorf("Unable to decode file output: %v", err)
	}
	return ioutil.NopCloser(bytes.NewReader(byts)), nil
}
