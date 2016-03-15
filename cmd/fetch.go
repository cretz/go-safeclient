package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"errors"
	"github.com/cretz/go-safeclient/client"
	"os"
	"io"
)

var fetchShared bool
var fetchToFile string
var fetchOffset int64
var fetchLength int64

var fetchCmd = &cobra.Command{
	Use:   "fetch [file path]",
	Short: "Fetch file contents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.GetFileInfo{
			FilePath: args[0],
			Shared: fetchShared,
			Offset: fetchOffset,
			Length: fetchLength,
		}
		rc, err := c.GetFile(info)
		if err != nil {
			log.Fatalf("Failed to read file: %v", err)
		}
		defer rc.Close()
		outFile := os.Stdout
		if fetchToFile != "" {
			if outFile, err = os.Create(fetchToFile); err != nil {
				log.Fatalf("Unable to create output file: %v", err)
			}
			defer outFile.Close()
		}
		if _, err = io.Copy(outFile, rc); err != nil {
			log.Fatalf("Unable to write to stdout: %v", err)
		}
		return nil
	},
}

func init() {
	fetchCmd.Flags().BoolVarP(&fetchShared, "shared", "s", false, "Use shared area for user/app")
	fetchCmd.Flags().StringVarP(&fetchToFile, "file", "f", "", "Write to file instead of stdout")
	fetchCmd.Flags().Int64VarP(&fetchOffset, "offset", "o", 0, "Offset to start writing from")
	fetchCmd.Flags().Int64VarP(&fetchLength, "length", "l", 0, "Amount of bytes to read")
	RootCmd.AddCommand(fetchCmd)
}
