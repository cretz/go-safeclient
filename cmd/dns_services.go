package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var dnsServicesCmd = &cobra.Command{
	Use:   "services [dns name]",
	Short: "Get all DNS services for the given name",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatal("One argument and only one argument for name required")
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
	},
}

func init() {
	dnsCmd.AddCommand(dnsServicesCmd)
}
