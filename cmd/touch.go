package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"errors"
	"github.com/cretz/go-safeclient/client"
)

var touchShared bool
var touchMetadata string

var touchCmd = &cobra.Command{
	Use:   "touch [file path]",
	Short: "Create empty file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.CreateFileInfo{
			FilePath: args[0],
			Shared: touchShared,
			Metadata: touchMetadata,
		}
		if err = c.CreateFile(info); err != nil {
			log.Fatalf("Failed to create file: %v", err)
		}
		return nil
	},
}

func init() {
	touchCmd.Flags().BoolVarP(&touchShared, "shared", "s", false, "Use shared area for user/app")
	touchCmd.Flags().StringVarP(&touchMetadata, "metadata", "m", "", "Metadata to store with the file")
	RootCmd.AddCommand(touchCmd)
}
