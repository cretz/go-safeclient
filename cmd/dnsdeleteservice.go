package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
)

var dnsDeleteServiceCmd = &cobra.Command{
	Use:   "dnsdeleteservice [dns name] [dns service]",
	Short: "Delete DNS service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Must have exactly two arguments for DNS name and service")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		if err = c.DNSDeleteService(args[0], args[1]); err != nil {
			log.Fatalf("Unable to delete service: %v", err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dnsDeleteServiceCmd)
}
