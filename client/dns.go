package client

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
)

// DNSNames gets all top level DNS names on the account. See https://maidsafe.readme.io/docs/dns-list-long-names for
// more info.
func (c *Client) DNSNames() ([]string, error) {
	ret := []string{}
	req := &Request{
		Path:         "/dns",
		Method:       "GET",
		JSONResponse: &ret,
	}
	if _, err := c.Do(req); err != nil {
		return nil, err
	}
	return ret, nil
}

// DNSServices gets all DNS services on the account for the given name. See
// https://maidsafe.readme.io/docs/dns-list-services for more information.
func (c *Client) DNSServices(name string) ([]string, error) {
	ret := []string{}
	req := &Request{
		// XXX: we don't care about santizing this URL, the server should (what if name does start with slash?)
		Path:         "/dns/" + url.QueryEscape(name),
		Method:       "GET",
		JSONResponse: &ret,
	}
	if _, err := c.Do(req); err != nil {
		return nil, err
	}
	return ret, nil
}

// DNSServiceDir gets directory detail for the given DNS name and service on the account. See
// https://maidsafe.readme.io/docs/dns-get-home-dir for more info.
func (c *Client) DNSServiceDir(name, service string) (DirResponse, error) {
	var resp DirResponse
	req := &Request{
		Path:         "/dns/" + url.QueryEscape(service) + "/" + url.QueryEscape(name),
		Method:       "GET",
		JSONResponse: &resp,
		DoNotEncrypt: true,
		DoNotAuth:    true,
	}
	_, err := c.Do(req)
	return resp, err
}

// DNSFileInfo are parameters for Client.DNSFile
type DNSFileInfo struct {
	// The DNS name this file is at
	Name string
	// The service name this file is at
	Service string
	// The path to the file
	FilePath string
	// The offset to read from
	Offset int64
	// The number of bytes to read. 0 means no limit.
	Length int64
}

// DNSFile is returned from the Client.DNSFile function
type DNSFile struct {
	// File info about the DNS file
	Info FileInfo
	// The mime type of the content from SAFE
	ContentType string
	// The contents of the file
	Body io.ReadCloser
}

// DNSFile fetches a public file from DNS. See https://maidsafe.readme.io/docs/dns-get-file-unauth for more information.
func (c *Client) DNSFile(df DNSFileInfo) (*DNSFile, error) {
	query := map[string][]string{"offset": []string{strconv.FormatInt(df.Offset, 10)}}
	if df.Length > 0 {
		query["length"] = []string{strconv.FormatInt(df.Length, 10)}
	}
	req := &Request{
		Path: "/dns/" + url.QueryEscape(df.Service) + "/" + url.QueryEscape(df.Name) +
			"/" + url.QueryEscape(df.FilePath),
		Method:       "GET",
		Query:        query,
		DoNotEncrypt: true,
		DoNotAuth:    true,
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return &DNSFile{
		Info: FileInfo{
			Name:       resp.Header.Get("file-name"),
			Size:       maybeHeaderInt64(resp.Header, "file-size"),
			CreatedOn:  Time(maybeHeaderInt64(resp.Header, "file-created-time")),
			ModifiedOn: Time(maybeHeaderInt64(resp.Header, "file-modified-time")),
			// XXX: Ug, this can be "undefined"
			Metadata: resp.Header.Get("file-metadata"),
		},
		ContentType: resp.Header.Get("Content-Type"),
		Body:        resp.Body,
	}, nil
}

func maybeHeaderInt64(header http.Header, name string) int64 {
	val, err := strconv.ParseInt(header.Get(name), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

// DNSRegisterInfo are parameters for Client.DNSRegister
type DNSRegisterInfo struct {
	// The DNS name to register
	Name string `json:"longName"`
	// The service name to register
	ServiceName string `json:"serviceName"`
	// The home directory path to register
	HomeDirPath string `json:"serviceHomeDirPath"`
	// Whether the path is shared or not
	Shared bool `json:"isPathShared"`
}

// DNSRegister registers a DNS top level name, service, and directory. This differs from DNSAddService because it
// internally calls DNSCreateName. See https://maidsafe.readme.io/docs/dns-register-service for more info.
func (c *Client) DNSRegister(dr DNSRegisterInfo) error {
	req := &Request{
		Path:     "/dns",
		Method:   "POST",
		JSONBody: dr,
	}
	_, err := c.Do(req)
	return err
}

// DNSCreateName creates a new top level DNS name for this account. See
// https://maidsafe.readme.io/docs/dns-create-long-name for more info.
func (c *Client) DNSCreateName(name string) error {
	req := &Request{
		Path:   "/dns/" + url.QueryEscape(name),
		Method: "POST",
	}
	_, err := c.Do(req)
	return err
}

// DNSAddServiceInfo is returned from the Client.DNSAddService
type DNSAddServiceInfo struct {
	// The DNS name to add the service to
	Name string `json:"longName"`
	// The service name to add
	ServiceName string `json:"serviceName"`
	// The directory to assign
	HomeDirPath string `json:"serviceHomeDirPath"`
	// Whether the path is shared
	Shared bool `json:"isPathShared"`
}

// DNSAddService registers a service and directory to an existing DNS top level name. This differs from DNSRegister
// because it expects the name to already exist. This is not documented, see https://maidsafe.atlassian.net/browse/CS-60
// for more info.
func (c *Client) DNSAddService(das DNSAddServiceInfo) error {
	// TODO: broken?
	req := &Request{
		Path:     "/dns",
		Method:   "PUT",
		JSONBody: das,
	}
	_, err := c.Do(req)
	return err
}

// DNSDeleteName deletes a top-level DNS name
func (c *Client) DNSDeleteName(name string) error {
	req := &Request{
		Path:   "/dns/" + url.QueryEscape(name),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}

// DNSDeleteService deletes a service from the given DNS name
func (c *Client) DNSDeleteService(name, service string) error {
	req := &Request{
		Path:   "/dns/" + url.QueryEscape(service) + "/" + url.QueryEscape(name),
		Method: "DELETE",
	}
	_, err := c.Do(req)
	return err
}
