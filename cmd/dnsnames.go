package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var dnsNamesCmd = &cobra.Command{
	Use:   "dnsnames",
	Short: "Get all DNS names",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return errors.New("No arguments expected for this command")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		names, err := c.DNSNames()
		if err != nil {
			log.Fatalf("Unable to list names: %v", err)
		}
		if len(names) > 0 {
			fmt.Print(strings.Join(names, "\n") + "\n")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dnsNamesCmd)
}
