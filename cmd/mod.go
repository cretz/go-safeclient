package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"errors"
	"github.com/cretz/go-safeclient/client"
)

var modShared bool
var modName string
var modMetadata string

var modCmd = &cobra.Command{
	Use:   "mod [file]",
	Short: "Change file name and/or metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		} else if (modName == "" && modMetadata == "") {
			return errors.New("Must provide at least name or metadata")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.ChangeFileInfo{
			FilePath: args[0],
			Shared: modShared,
			NewName: modName,
			Metadata: modMetadata,
		}
		if err = c.ChangeFile(info); err != nil {
			log.Fatalf("Failed to change file: %v", err)
		}
		return nil
	},
}

func init() {
	modCmd.Flags().BoolVarP(&modShared, "shared", "s", false, "Use shared area for user/app")
	modCmd.Flags().StringVarP(&modName, "name", "n", "", "The name to change to")
	modCmd.Flags().StringVarP(&modMetadata, "metadata", "m", "", "Metadata to update")
	RootCmd.AddCommand(modCmd)
}
