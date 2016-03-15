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
	"strings"
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

type APIError struct {
	HttpResponse *http.Response
}

func NewAPIError(resp *http.Response) *APIError {
	return &APIError{HttpResponse: resp}
}

func (a *APIError) Error() string {
	body := a.HttpResponse.Body
	defer body.Close()
	bodyByts, _ := ioutil.ReadAll(a.HttpResponse.Body)
	// Put the body back if they want to use it
	a.HttpResponse.Body = ioutil.NopCloser(bytes.NewReader(bodyByts))
	return fmt.Sprintf("Server error %v: %v", a.HttpResponse.StatusCode, string(bodyByts))
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
	RawBody      []byte
	Query        url.Values
	DoNotEncrypt bool
	JSONResponse interface{}
	DoNotAuth    bool
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
	if err := c.sanitizeResponse(httpResp, !req.DoNotEncrypt, req.JSONResponse); err != nil {
		return nil, err
	}
	return httpResp, err
}

func (c *Client) buildRequest(req *ClientRequest) (*http.Request, error) {
	fullURL, err := url.Parse(c.Conf.LauncherBaseURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid launcher base URL: %v", err)
	}
	// Need path escaping as given to us by the user
	fullURL.Path, err = url.QueryUnescape(req.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to unescape given path: %v", err)
	}
	fullURL.RawPath = req.Path
	fullURL.RawQuery = req.Query.Encode()
	httpReq := &http.Request{
		Method: req.Method,
		URL:    fullURL,
		Header: map[string][]string{},
	}

	// If there is a body, handle it
	if req.JSONBody != nil {
		req.RawBody, err = json.Marshal(req.JSONBody)
		if err != nil {
			return nil, fmt.Errorf("Unable to JSON marshal body: %v", err)
		}
	}
	if req.RawBody != nil {
		// Encrypt if necessary
		if req.DoNotEncrypt {
			httpReq.Body = ioutil.NopCloser(bytes.NewReader(req.RawBody))
			httpReq.ContentLength = int64(len(req.RawBody))
			if req.JSONBody != nil {
				httpReq.Header["Content-Type"] = []string{"application/json"}
			} else {
				httpReq.Header["Content-Type"] = []string{"text/plain"}
			}
		} else {
			encrypted := base64.StdEncoding.EncodeToString(c.encrypt(req.RawBody))
			httpReq.Body = ioutil.NopCloser(strings.NewReader(encrypted))
			httpReq.ContentLength = int64(len(encrypted))
			httpReq.Header["Content-Type"] = []string{"text/plain"}
		}
	}

	// If there is a token, use it as the bearer token
	if c.Conf.Token != "" && !req.DoNotAuth {
		httpReq.Header["authorization"] = []string{"Bearer " + c.Conf.Token}
	}

	// Encrypt the query string if necessary
	if !req.DoNotEncrypt && httpReq.URL.RawQuery != "" {
		out := c.encrypt([]byte(httpReq.URL.RawQuery))
		httpReq.URL.RawQuery = url.QueryEscape(base64.StdEncoding.EncodeToString(out))
	}
	return httpReq, nil
}

var doNotDecryptResponses = map[int]string{
	200: "OK",
	202: "Accepted",
	500: "Server Error",
}

func (c *Client) sanitizeResponse(resp *http.Response, decrypt bool, jsonResponse interface{}) error {
	if decrypt {
		body := resp.Body
		defer body.Close()
		encryptedBase64d, err := ioutil.ReadAll(body)
		if err != nil {
			return fmt.Errorf("Unable to read response: %v", err)
		}
		// As a special case, there can be some values which we don't mess with
		if v, _ := doNotDecryptResponses[resp.StatusCode]; string(encryptedBase64d) == v {
			resp.Body = ioutil.NopCloser(bytes.NewReader(encryptedBase64d))
		} else if len(encryptedBase64d) > 0 {
			encrypted := make([]byte, base64.StdEncoding.DecodedLen(len(encryptedBase64d)))
			if n, err := base64.StdEncoding.Decode(encrypted, encryptedBase64d); err != nil {
				// When we can't base 64 decode it, we just fall back to the normal
				resp.Body = ioutil.NopCloser(bytes.NewReader(encryptedBase64d))
			} else {
				encrypted = encrypted[:n]
				decrypted, err := c.decrypt(encrypted)
				if err != nil {
					return err
				}
				resp.Body = ioutil.NopCloser(bytes.NewReader(decrypted))
			}
		} else {
			resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
		}
	}
	// We consider non-200 as a failure
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewAPIError(resp)
	}
	if jsonResponse != nil {
		// Go ahead and take entire body and unmarshal it (and then put it back)
		body := resp.Body
		defer body.Close()
		jsonByts, err := ioutil.ReadAll(body)
		if err != nil {
			return fmt.Errorf("Unable to read body: %v", err)
		}
		if len(jsonByts) > 0 {
			if err = json.Unmarshal(jsonByts, jsonResponse); err != nil {
				return fmt.Errorf("Unable to unmarshal JSON: %v", err)
			}
		}
		resp.Body = ioutil.NopCloser(bytes.NewReader([]byte{}))
	}
	return nil
}

func (c *Client) encrypt(in []byte) []byte {
	var nonce [24]byte
	copy(nonce[:], c.Conf.Nonce)
	var sharedKey [32]byte
	copy(sharedKey[:], c.Conf.SharedKey)
	return secretbox.Seal([]byte{}, in, &nonce, &sharedKey)
}

func (c *Client) decrypt(in []byte) ([]byte, error) {
	var nonce [24]byte
	copy(nonce[:], c.Conf.Nonce)
	var sharedKey [32]byte
	copy(sharedKey[:], c.Conf.SharedKey)
	out, ok := secretbox.Open([]byte{}, in, &nonce, &sharedKey)
	if !ok {
		return nil, errors.New("Failed to decrypt")
	}
	return out, nil
}
