package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var dnsServiceDirCmd = &cobra.Command{
	Use:   "dnsservicedir [dns name] [dns service]",
	Short: "Get DNS service dir for the given name and service",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Must have exactly two arguments for DNS name and service")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		dir, err := c.DNSServiceDir(args[0], args[1])
		if err != nil {
			log.Fatalf("Unable to get dir: %v", err)
		}
		writeDirResponseTable(os.Stdout, dir)
		return nil
	},
}

func init() {
	RootCmd.AddCommand(dnsServiceDirCmd)
}
