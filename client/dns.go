package client

import (
	"strings"
	"io"
	"net/http"
	"strconv"
)

func (c *Client) DNSNames() ([]string, error) {
	ret := []string{}
	req := &ClientRequest{
		Path: "/dns",
		Method: "GET",
		JSONResponse: &ret,
	}
	if _, err := c.Do(req); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) DNSServices(name string) ([]string, error) {
	ret := []string{}
	req := &ClientRequest{
		// XXX: we don't care about santizing this URL, the server should (what if name does start with slash?)
		Path: "/dns/" + name,
		Method: "GET",
		JSONResponse: &ret,
	}
	if _, err := c.Do(req); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *Client) DNSServiceDir(name, service string) (DirResponse, error) {
	var resp DirResponse
	req := &ClientRequest{
		Path: "/dns/" + service + "/" + name + "/",
		Method: "GET",
		JSONResponse: &resp,
		DoNotEncrypt: true,
		DoNotAuth: true,
	}
	_, err := c.Do(req)
	return resp, err
}

type DNSFile struct {
	Info FileInfo
	ContentType string
	Body io.ReadCloser
}

func (c *Client) DNSFile(name, service, path string) (*DNSFile, error) {
	req := &ClientRequest{
		Path: "/dns/" + service + "/" + name + "/" + strings.TrimLeft(path, "/"),
		Method: "GET",
		DoNotEncrypt: true,
		DoNotAuth: true,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return &DNSFile{
		Info: FileInfo{
			Name: resp.Header.Get("file-name"),
			Size: maybeHeaderInt64(resp.Header, "file-size"),
			CreatedOn: Time(maybeHeaderInt64(resp.Header, "file-created-time")),
			ModifiedOn: Time(maybeHeaderInt64(resp.Header, "file-modified-time")),
			// XXX: Ug, this can be "undefined"
			Metadata: resp.Header.Get("file-metadata"),
		},
		ContentType: resp.Header.Get("Content-Type"),
		Body: resp.Body,
	}, nil
}

func maybeHeaderInt64(header http.Header, name string) int64 {
	val, err := strconv.ParseInt(header.Get(name), 10, 64)
	if err != nil {
		return 0
	}
	return val
}