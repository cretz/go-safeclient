package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
)

var dnsFileOutFile = ""
var dnsFileOffset int64
var dnsFileLength int64

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
		info := client.DNSFileInfo{
			Name:     args[0],
			Service:  args[1],
			FilePath: args[2],
			Offset:   dnsFileOffset,
			Length:   dnsFileLength,
		}
		file, err := c.DNSFile(info)
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
	dnsFileCmd.Flags().StringVarP(&dnsFileOutFile, "file", "f", "", "Output to file instead of stdout")
	dnsFileCmd.Flags().Int64VarP(&dnsFileOffset, "offset", "o", 0, "Offset to start writing from")
	dnsFileCmd.Flags().Int64VarP(&dnsFileLength, "length", "l", 0, "Amount of bytes to read")
	RootCmd.AddCommand(dnsFileCmd)
}
