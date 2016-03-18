package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"os"
)

// RootCmd is the root command that the CLI runs
var RootCmd = &cobra.Command{
	Use:   "go-safeclient",
	Short: "go-safeclient is a simple app that shows how to do simple things with the safe launcher",
}

var cfgFile = ""
var verbose = false

var app = client.AuthAppInfo{
	Name:    "SAFE Client CLI",
	ID:      "go-safeclient.cretz.github.com",
	Version: "0.0.1",
	Vendor:  "cretz",
	//Name:    "Maidsafe Demo",
	//ID:      "demo.maidsafe.net",
	//Version: "0.2.1",
	//Vendor:  "MaidSafe",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "conf.json", "config file")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "show debug output")
}

func getClient() (*client.Client, error) {
	// Try to load the config first
	var conf client.Conf
	if _, err := os.Stat(cfgFile); !os.IsNotExist(err) {
		byts, err := ioutil.ReadFile(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("Unable to read conf: %v", err)
		}
		// JSON unmarshal into the conf
		if err := json.Unmarshal(byts, &conf); err != nil {
			return nil, fmt.Errorf("Invalid conf: %v", err)
		}
	}

	// Create the client and ensure it is authed
	c := client.NewClient(conf)
	if verbose {
		c.Logger = log.New(os.Stderr, "", log.LstdFlags)
	}
	authInfo := client.AuthInfo{App: app, Permissions: []string{client.AuthPermSafeDriveAccess}}
	if err := c.EnsureAuthed(authInfo); err != nil {
		return nil, err
	}

	// Persist the config
	persistBytes, err := json.Marshal(c.Conf)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(cfgFile, persistBytes, 0600); err != nil {
		return nil, fmt.Errorf("Unable to write config at %v: %v", cfgFile, err)
	}
	return c, nil
}
