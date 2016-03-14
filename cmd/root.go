package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var RootCmd = &cobra.Command{
	Use:   "go-safeclient",
	Short: "go-safeclient is a simple app that shows how to do simple things with the safe launcher",
}

var cfgFile = ""

var app = client.AuthApp{
	Name:    "SAFE Client CLI",
	ID:      "go-safeclient.cretz.github.com",
	Version: "0.0.1",
	Vendor:  "cretz",
}

func init() {
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "conf.json", "config file")
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
	client := client.NewClient(conf)
	if err := client.EnsureAuthed(app); err != nil {
		return nil, err
	}

	// Persist the config
	persistBytes, err := json.Marshal(client.Conf)
	if err != nil {
		return nil, err
	}
	if err := ioutil.WriteFile(cfgFile, persistBytes, 600); err != nil {
		return nil, fmt.Errorf("Unable to write config at %v: %v", cfgFile, err)
	}
	return client, nil
}
