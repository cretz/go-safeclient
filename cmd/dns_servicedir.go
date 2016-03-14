package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
)

var dnsServiceDirCmd = &cobra.Command{
	Use:   "servicedir [dns name] [dns service]",
	Short: "Get DNS service dir for the given name and service",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			log.Fatal("Must have exactly two arguments for DNS name and service")
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
	},
}

func init() {
	dnsCmd.AddCommand(dnsServiceDirCmd)
}
