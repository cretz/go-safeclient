package client

import (
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/box"
)

// AuthInfo is used by Client.Auth and Client.EnsureAuthed to authenticate an app
type AuthInfo struct {
	// This is the required app information
	App AuthAppInfo `json:"app"`
	// Set of permissions that this app will register with. See AuthPerm* constants.
	Permissions []string `json:"permissions"`
	// The public key to use when authorizing. Leave blank to generate at runtime.
	PublicKey []byte `json:"publicKey"`
	// The private key to use when authorizing. Leave blank to generate at runtime.
	PrivateKey []byte `json:"-"`
	// The nonce to use when authorizing. Leave blank to generate at runtime.
	Nonce []byte `json:"nonce"`
}

// AuthAppInfo is the app identification detail
type AuthAppInfo struct {
	// The name of the app that will appear in the SAFE launcher
	Name string `json:"name"`
	// The ID of the app
	ID string `json:"id"`
	// The app version
	Version string `json:"version"`
	// The app vendor name
	Vendor string `json:"vendor"`
}

// AuthResult is the result of a successful authentication request
type AuthResult struct {
	// The original request
	Request AuthInfo
	// The JSON web token for this app to use
	Token string
	// The shared key to use when encrypting/decrypting to/from the SAFE launcher
	SharedKey []byte
	// The nonce to use when encrypting/decrypting to/from the SAFE launcher
	Nonce []byte
}

// ErrAuthDenied is the error returned when the user denied access to the application during Client.Auth and
// Client.EnsureAuthed
var ErrAuthDenied = errors.New("Auth denied")

const (
	// AuthPermSafeDriveAccess is a permission to have access to the shared SAFE drive
	AuthPermSafeDriveAccess = "SAFE_DRIVE_ACCESS"
)

// Auth authenticates with the SAFE launcher which prompts the user to accept it. If the AuthInfo.PublicKey is not
// provided, it will be generated automatically along with the PrivateKey and Nonce. See
// https://maidsafe.readme.io/docs/auth for more info.
func (c *Client) Auth(ai AuthInfo) (AuthResult, error) {
	res := AuthResult{}
	withKey := ai
	if len(withKey.PublicKey) == 0 {
		// Go ahead and create a key here for our use
		pubKey, privKey, err := box.GenerateKey(rand.Reader)
		if err != nil {
			return res, fmt.Errorf("Unable to generate key: %v", err)
		}
		withKey.PublicKey = pubKey[:]
		withKey.PrivateKey = privKey[:]
		withKey.Nonce = make([]byte, 24)
		if _, err = rand.Read(withKey.Nonce); err != nil {
			return res, fmt.Errorf("Unable to generate nonce: %v", err)
		}
	}
	// XXX: Have to have an empty permissions object sadly
	if withKey.Permissions == nil {
		withKey.Permissions = []string{}
	}
	authResp := &authResponse{}
	req := &Request{
		Path:         "/auth",
		Method:       "POST",
		JSONBody:     withKey,
		DoNotEncrypt: true,
		JSONResponse: authResp,
	}
	if _, err := c.Do(req); err != nil {
		if v, ok := err.(*APIError); ok && v.HTTPResponse.StatusCode == 401 {
			return res, ErrAuthDenied
		}
		return res, err
	}
	// Build response from given key
	res.Request = withKey
	if err := withKey.applyResponseToResult(authResp, &res); err != nil {
		return res, err
	}
	return res, nil
}

// IsValidToken checks if the current Client.Conf information is valid to access the API. See
// https://maidsafe.readme.io/docs/is-token-valid for more information.
func (c *Client) IsValidToken() (bool, error) {
	if c.Conf.Token == "" {
		return false, nil
	}
	req := &Request{
		Path:         "/auth",
		Method:       "GET",
		DoNotEncrypt: true,
	}
	resp, err := c.Do(req)
	if err != nil {
		if v, ok := err.(*APIError); ok && v.HTTPResponse.StatusCode == 401 {
			return false, nil
		}
		return false, err
	}
	return resp.StatusCode == 200, nil
}

// EnsureAuthed checks Client.IsValidToken and if false runs Client.Auth after clearing existing information from
// Client.Conf. It autopopulates Client.Conf.Token, Client.Conf.SharedKey, and Client.Conf.Nonce if the current
// Client.Conf is not valid.
func (c *Client) EnsureAuthed(ai AuthInfo) error {
	if valid, err := c.IsValidToken(); err != nil {
		return err
	} else if !valid {
		c.Conf.Token = ""
		c.Conf.SharedKey = nil
		c.Conf.Nonce = nil
		resp, err := c.Auth(ai)
		if err != nil {
			return err
		}
		c.Conf.Token = resp.Token
		c.Conf.SharedKey = resp.SharedKey
		c.Conf.Nonce = resp.Nonce
	}
	return nil
}

type authResponse struct {
	Token        string `json:"token"`
	EncryptedKey []byte `json:"encryptedKey"`
	PublicKey    []byte `json:"publicKey"`
}

func (a AuthInfo) applyResponseToResult(resp *authResponse, res *AuthResult) error {
	var nonce [24]byte
	copy(nonce[:], a.Nonce)
	var peerPubKey, privKey [32]byte
	copy(peerPubKey[:], resp.PublicKey)
	copy(privKey[:], a.PrivateKey)
	out, ok := box.Open([]byte{}, resp.EncryptedKey, &nonce, &peerPubKey, &privKey)
	if !ok {
		return errors.New("Unable to decrypt auth key")
	}
	res.Token = resp.Token
	res.SharedKey = out[:32]
	res.Nonce = out[32:]
	return nil
}
