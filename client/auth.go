package client

import (
	"crypto/rand"
	"errors"
	"fmt"
	"golang.org/x/crypto/nacl/box"
)

type AuthRequest struct {
	App         AuthApp  `json:"app"`
	Permissions []string `json:"permissions"`
	PublicKey   []byte   `json:"publicKey"`
	PrivateKey  []byte   `json:"-"`
	Nonce       []byte   `json:"nonce"`
}

type AuthApp struct {
	Name    string `json:"name"`
	ID      string `json:"id"`
	Version string `json:"version"`
	Vendor  string `json:"vendor"`
}

type AuthResult struct {
	Request   AuthRequest
	Token     string
	SharedKey []byte
	Nonce     []byte
}

var AuthDeniedError = errors.New("Auth denied")

func (a AuthRequest) Auth(client *Client) (AuthResult, error) {
	res := AuthResult{}
	withKey := a
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
	withKey.Permissions = []string{}
	authResp := &authResponse{}
	req := &ClientRequest{
		Path:         "/auth",
		Method:       "POST",
		JSONBody:     withKey,
		DoNotEncrypt: true,
		JSONResponse: authResp,
	}
	if _, err := client.Do(req); err != nil {
		if v, ok := err.(*APIError); ok && v.HttpResponse.StatusCode == 401 {
			return res, AuthDeniedError
		} else {
			return res, err
		}
	}
	// Build response from given key
	res.Request = withKey
	if err := withKey.applyResponseToResult(authResp, &res); err != nil {
		return res, err
	}
	return res, nil
}

func (c *Client) IsValidToken() (bool, error) {
	if c.Conf.Token == "" {
		return false, nil
	}
	req := &ClientRequest{
		Path:         "/auth",
		Method:       "GET",
		DoNotEncrypt: true,
	}
	resp, err := c.Do(req)
	if err != nil {
		if v, ok := err.(*APIError); ok && v.HttpResponse.StatusCode == 401 {
			return false, nil
		} else {
			return false, err
		}
	}
	return resp.StatusCode == 200, nil
}

func (c *Client) EnsureAuthed(app AuthApp, permissions ...string) error {
	if valid, err := c.IsValidToken(); err != nil {
		return err
	} else if !valid {
		c.Conf.Token = ""
		c.Conf.SharedKey = nil
		c.Conf.Nonce = nil
		req := AuthRequest{
			App:         app,
			Permissions: permissions,
		}
		resp, err := req.Auth(c)
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

func (a AuthRequest) applyResponseToResult(resp *authResponse, res *AuthResult) error {
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
