package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"io"
)

var dnsFileCmd = &cobra.Command{
	Use:   "file [dns name] [dns service] [file path]",
	Short: "Get file at specific path for DNS name and service name",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 3 {
			log.Fatal("Must have exactly three arguments for DNS name, service name, and file path")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		file, err := c.DNSFile(args[0], args[1], args[2])
		if err != nil {
			log.Fatalf("Unable to get file: %v", err)
		}
		defer file.Body.Close()
		if _, err := io.Copy(os.Stdout, file.Body); err != nil {
			log.Fatalf("Unable to copy to output: %v", err)
		}
	},
}

func init() {
	dnsCmd.AddCommand(dnsFileCmd)
}
