package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"os"
	"io"
	"errors"
)

var dnsFileOutFile = ""

var dnsFileCmd = &cobra.Command{
	Use:   "dnsfile [dns name] [dns service] [file path]",
	Short: "Get file at specific path for DNS name and service name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return errors.New("Must have exactly three arguments for DNS name, service name, and file path")
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
		out := os.Stdout
		if dnsFileOutFile != "" {
			if out, err = os.Create(dnsFileOutFile); err != nil {
				log.Fatalf("Unable to create out file: %v", err)
			}
			defer out.Close()
		}
		if _, err := io.Copy(os.Stdout, file.Body); err != nil {
			log.Fatalf("Unable to copy to output: %v", err)
		}
		return nil
	},
}

func init() {
	dnsFileCmd.Flags().StringVarP(&dnsFileOutFile, "outfile", "o", "", "Output to file instead of stdout")
	RootCmd.AddCommand(dnsFileCmd)
}
