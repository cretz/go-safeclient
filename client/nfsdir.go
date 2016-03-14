package client

import (
	"strconv"
	"errors"
	"io/ioutil"
	"fmt"
	"net/url"
)

func (c *Client) GetDir(dirPath string, shared bool) error {
	req := &ClientRequest{
		Path: "/nfs/directory/" + url.QueryEscape(dirPath) + "/" + strconv.FormatBool(shared),
		Method: "GET",
		DoNotAuth: true,
	}
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	byts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("Unable to read body: %v", err)
	}
	fmt.Printf("BODY: %v\n", string(byts))
	return errors.New("NO!")
}
