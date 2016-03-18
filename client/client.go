package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/secretbox"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

// Conf is the configuration for a client. It is built to marshal to JSON which can be stored in a file for reuse.
type Conf struct {
	// The LauncherBaseURL. If this is not present it is defaulted to http://localhost:8100/ in NewClient.
	LauncherBaseURL string `json:"launcherServer,omitempty"`
	// The JSON web token to use to authenticate on authenticated calls. This is required for authenticated calls.
	// It is autopopulated by Client.EnsureAuthed if invalid.
	Token string `json:"token,omitempty"`
	// The shared key for encryption/decryption. This is required for encrypted calls. It is autopopulated by
	// Client.EnsureAuthed if invalid.
	SharedKey []byte `json:"sharedKey,omitempty"`
	// The nonce for encryption/decryption. This is required for encrypted calls. It is autopopulated by
	// Client.EnsureAuthed if invalid.
	Nonce []byte `json:"nonce,omitempty"`
}

// RequestBuilder is a function type that is used to build an HTTP request from a client request.
type RequestBuilder func(c *Client, req *Request) (*http.Request, error)

// ResponseHandler is a function type that is used to alter an HTTP response as needed. If decrypt is true,
// implementers are expected to decrypt the response body and place it back on the response. If jsonResponse is not nil,
// implementers are expected to unmarshal the JSON into the object.
type ResponseHandler func(c *Client, resp *http.Response, decrypt bool, jsonResponse interface{}) error

// Client for accessing SAFE. While this can be constructed manually, NewClient populates some of these values.
type Client struct {
	// The configuration for the client including auth and encryption information
	Conf Conf
	// The HTTP client to use. This must be present. NewClient populates this with http.DefaultClient by default.
	HTTPClient *http.Client
	// A custom request builder. If not set it will default to the normal builder. NewClient populates this with the
	// default value so developers can override this value as a proxy and call the existing one if necessary.
	RequestBuilder RequestBuilder
	// A custom response handler. If not set it will default to the normal builder. NewClient populates this with
	// the default value so developers can override this value as a proxy and call the existing one if necessary.
	ResponseHandler ResponseHandler
	// If present, debug logs will be logged here
	Logger *log.Logger
}

// APIError represents a server-side API error on non-2xx responses
type APIError struct {
	// The response that triggered this error
	HTTPResponse *http.Response
}

// NewAPIError creates an APIError object from the given HTTP response
func NewAPIError(resp *http.Response) *APIError {
	return &APIError{HTTPResponse: resp}
}

// Error gives details about the error including HTTP status code and the entire body
func (a *APIError) Error() string {
	body := a.HTTPResponse.Body
	defer body.Close()
	bodyByts, _ := ioutil.ReadAll(a.HTTPResponse.Body)
	// Put the body back if they want to use it
	a.HTTPResponse.Body = ioutil.NopCloser(bytes.NewReader(bodyByts))
	return fmt.Sprintf("Server error %v: %v", a.HTTPResponse.StatusCode, string(bodyByts))
}

// NewClient constructs a new client with the given conf.
func NewClient(conf Conf) *Client {
	newConf := conf
	// Default to localhost base url
	if newConf.LauncherBaseURL == "" {
		newConf.LauncherBaseURL = "http://localhost:8100/"
	}
	return &Client{
		Conf:       newConf,
		HTTPClient: http.DefaultClient,
		RequestBuilder: func(c *Client, req *Request) (*http.Request, error) {
			return c.buildRequest(req)
		},
		ResponseHandler: func(c *Client, resp *http.Response, decrypt bool, jsonResponse interface{}) error {
			return c.handleResponse(resp, decrypt, jsonResponse)
		},
	}
}

// Request is a representation of a request to SAFE
type Request struct {
	// The URL path to call not including the host information. Note, if there are things inside of the path that
	// should be URL encoded, they must be encoded before calling this. If this value does not start with a slash,
	// it is treated as though it does.
	Path string
	// The HTTP method to use for the call
	Method string
	// If not nil, this value will be JSON marshalled into the body overriding any value that may be in RawBody
	JSONBody interface{}
	// If not nil and JSONBody is nil, this will be used as the HTTP request body. NOTE: an empty array is still
	// considered a body of content length 0.
	RawBody []byte
	// The query values to send in the request
	Query url.Values
	// If true, the request body and query params will NOT be encrypted and the response body will NOT be decrypted
	DoNotEncrypt bool
	// If not nil, successful responses will JSON-unmarshalled into this object
	JSONResponse interface{}
	// If true, the request will not be authenticated with Client.Conf.Token
	DoNotAuth bool
}

// Do makes the HTTP call to SAFE. Errors can be anything during the request or any non-2xx response.
func (c *Client) Do(req *Request) (*http.Response, error) {
	var httpReq *http.Request
	var err error
	if c.RequestBuilder != nil {
		httpReq, err = c.RequestBuilder(c, req)
	} else {
		httpReq, err = c.buildRequest(req)
	}
	if err != nil {
		return nil, err
	}
	httpResp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	if c.ResponseHandler != nil {
		err = c.ResponseHandler(c, httpResp, !req.DoNotEncrypt, req.JSONResponse)
	} else {
		err = c.handleResponse(httpResp, !req.DoNotEncrypt, req.JSONResponse)
	}
	if err != nil {
		return nil, err
	}
	return httpResp, err
}

func (c *Client) buildRequest(req *Request) (*http.Request, error) {
	fullURL, err := url.Parse(c.Conf.LauncherBaseURL)
	if err != nil {
		return nil, fmt.Errorf("Invalid launcher base URL: %v", err)
	}
	// Need path escaping as given to us by the user
	unescapedPath, err := url.QueryUnescape(req.Path)
	if err != nil {
		return nil, fmt.Errorf("Unable to unescape given path: %v", err)
	}
	fullURL.Path = strings.TrimSuffix(fullURL.Path, "/") + "/" + strings.TrimPrefix(unescapedPath, "/")
	fullURL.RawPath = req.Path
	fullURL.RawQuery = req.Query.Encode()
	httpReq := &http.Request{
		Method: req.Method,
		URL:    fullURL,
		Header: map[string][]string{},
	}
	if c.Logger != nil {
		c.Logger.Printf("Calling %v %v", httpReq.Method, httpReq.URL)
	}

	// If there is a body, handle it
	if req.JSONBody != nil {
		req.RawBody, err = json.Marshal(req.JSONBody)
		if err != nil {
			return nil, fmt.Errorf("Unable to JSON marshal body: %v", err)
		}
	}
	if req.RawBody != nil {
		if c.Logger != nil {
			c.Logger.Printf("REQ BODY: %v", string(req.RawBody))
		}
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

// Unfortunately simple status word responses are not encrypted which means we have to check it
// TODO: stop doing this and instead make it a per-call check. Otherwise there is a bug on a file that may be the value
// "8" (which is "OK" base 64 encoded).
var doNotDecryptResponses = map[int]string{
	200: "OK",
	202: "Accepted",
	500: "Server Error",
}

func (c *Client) handleResponse(resp *http.Response, decrypt bool, jsonResponse interface{}) error {
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
	if c.Logger != nil {
		defer resp.Body.Close()
		byts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		c.Logger.Printf("RESP BODY: %v", string(byts))
		resp.Body = ioutil.NopCloser(bytes.NewReader(byts))
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
