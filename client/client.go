package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"io/ioutil"
	"net/http"
	"net/url"
)

type Conf struct {
	LauncherBaseURL string `json:"launcherServer,omitempty"`
	Token           string `json:"token,omitempty"`
	SharedKey       []byte `json:"sharedKey,omitempty"`
	Nonce           []byte `json:"nonce,omitempty"`
}

type Client struct {
	Conf       Conf
	HttpClient *http.Client
}

type ClientResponse struct {
	// May be empty
	FileName string
	Body     []byte
}

type APIError struct {
	HttpResponse *http.Response
}

func NewAPIError(resp *http.Response) *APIError {
	return &APIError{HttpResponse: resp}
}

func (a *APIError) Error() string {
	return fmt.Sprintf("Server error, code: %v", a.HttpResponse.StatusCode)
}

func NewClient(conf Conf) *Client {
	newConf := conf
	// Default to localhost base url
	if newConf.LauncherBaseURL == "" {
		newConf.LauncherBaseURL = "http://localhost:8100/"
	}
	return &Client{Conf: newConf, HttpClient: http.DefaultClient}
}

type ClientRequest struct {
	Path         string
	Method       string
	JSONBody     interface{}
	Query        url.Values
	DoNotEncrypt bool
}

func (c *Client) Do(req *ClientRequest) (*http.Response, error) {
	httpReq, err := c.buildRequest(req)
	if err != nil {
		return nil, err
	}
	httpResp, err := c.HttpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if err := c.sanitizeResponse(httpResp, !req.DoNotEncrypt); err != nil {
		return nil, err
	}
	return httpResp, err
}

func (c *Client) buildRequest(req *ClientRequest) (*http.Request, error) {
	fullURL, err := url.Parse(c.Conf.LauncherBaseURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid launcher base URL: %v", err)
	}
	fullURL.Path = req.Path
	fullURL.RawQuery = req.Query.Encode()
	httpReq := &http.Request{
		Method: req.Method,
		URL:    fullURL,
		Header: map[string][]string{},
	}

	// If there is a body, json marshal it
	if req.JSONBody != nil {
		byts, err := json.Marshal(req.JSONBody)
		if err != nil {
			return nil, fmt.Errorf("Unable to JSON marshal body: %v", err)
		}
		// Encrypt if necessary
		if req.DoNotEncrypt {
			httpReq.Body = ioutil.NopCloser(bytes.NewBuffer(byts))
			httpReq.ContentLength = int64(len(byts))
			httpReq.Header["Content-Type"] = []string{"application/json"}
		} else {
			out := c.encrypt(byts)
			httpReq.Body = ioutil.NopCloser(bytes.NewBuffer(out))
			httpReq.ContentLength = int64(len(out))
			httpReq.Header["Content-Type"] = []string{"text/plain"}
		}
	}

	// If there is a token, use it as the bearer token
	if c.Conf.Token != "" {
		httpReq.Header["Authorization"] = []string{"Bearer " + c.Conf.Token}
	}

	// Encrypt the query string if necessary
	if !req.DoNotEncrypt && httpReq.URL.RawQuery != "" {
		out := c.encrypt([]byte(httpReq.URL.RawQuery))
		httpReq.URL.RawQuery = url.QueryEscape(base64.StdEncoding.EncodeToString(out))
	}
	return httpReq, nil
}

func (c *Client) sanitizeResponse(resp *http.Response, decrypt bool) error {
	if decrypt {
		body := resp.Body
		defer body.Close()
		encryptedBase64d, err := ioutil.ReadAll(body)
		if err != nil {
			return fmt.Errorf("Unable to read response: %v", err)
		}
		if len(encryptedBase64d) > 0 {
			encrypted := []byte{}
			if _, err := base64.StdEncoding.Decode(encrypted, encryptedBase64d); err != nil {
				return fmt.Errorf("Unable to decode base64 HTTP response: %v", err)
			}
			decrypted, err := c.decrypt(encrypted)
			if err != nil {
				return err
			}
			resp.Body = ioutil.NopCloser(bytes.NewBuffer(decrypted))
		} else {
			resp.Body = ioutil.NopCloser(bytes.NewBuffer([]byte{}))
		}
	}
	// We consider non-200 as a failure
	if resp.StatusCode != 200 {
		return NewAPIError(resp)
	}
	return nil
}

func (c *Client) encrypt(in []byte) []byte {
	var nonce [24]byte
	// TODO: This nonce is suppose to change regularly! Ask maidsafe devs
	// why the demo app's isn't
	copy(nonce[:], c.Conf.Nonce)
	var sharedKey [32]byte
	copy(sharedKey[:], c.Conf.SharedKey)
	out := []byte{}
	secretbox.Seal(out, in, &nonce, &sharedKey)
	return out
}

func (c *Client) decrypt(in []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], c.Conf.Nonce)
	var sharedKey [32]byte
	copy(sharedKey[:], c.Conf.SharedKey)
	out := []byte{}
	if _, ok := secretbox.Open(out, in, &nonce, &sharedKey); !ok {
		return nil, errors.New("Failed to decrypt")
	}
	return out, nil
}
