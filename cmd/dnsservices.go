package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var dnsServicesCmd = &cobra.Command{
	Use:   "dnsservices [dns name]",
	Short: "Get all DNS services for the given name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One argument and only one argument for name required")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		services, err := c.DNSServices(args[0])
		if err != nil {
			log.Fatalf("Unable to list services: %v", err)
		}
		if len(services) > 0 {
			fmt.Print(strings.Join(services, "\n") + "\n")
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dnsServicesCmd)
}
