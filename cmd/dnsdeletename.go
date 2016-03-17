package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
)

var dnsDeleteNameCmd = &cobra.Command{
	Use:   "dnsdeletename [name]",
	Short: "Delete DNS name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("Must have exactly 1 argument for name")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		if err = c.DNSDeleteName(args[0]); err != nil {
			log.Fatalf("Unable to delete name: %v", err)
		}
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dnsDeleteNameCmd)
}
