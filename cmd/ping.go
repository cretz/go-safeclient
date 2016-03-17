package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Do simple ping to make sure app is registered",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("No arguments expected for this command")
		}
		if c, err := getClient(); err != nil {
			log.Fatalf("Ping failed: %v", err)
		} else if ok, validErr := c.IsValidToken(); validErr != nil {
			log.Fatalf("Ping failed: %v", validErr)
		} else if !ok {
			log.Fatalf("Ping failed: Invalid token")
		}
		fmt.Println("Ping successful")
		return nil
	},
}

func init() {
	RootCmd.AddCommand(pingCmd)
}
