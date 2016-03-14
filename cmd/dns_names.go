package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"strings"
)

var dnsNamesCmd = &cobra.Command{
	Use:   "names",
	Short: "Get all DNS names",
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}

func init() {
	dnsCmd.AddCommand(dnsNamesCmd)
}
