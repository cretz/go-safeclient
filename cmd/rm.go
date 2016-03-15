package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var rmShared bool

var rmCmd = &cobra.Command{
	Use:   "rm [file]",
	Short: "Delete file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.DeleteFileInfo{
			FilePath: args[0],
			Shared:   rmShared,
		}
		if err = c.DeleteFile(info); err != nil {
			log.Fatalf("Failed to delete file: %v", err)
		}
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmShared, "shared", "s", false, "Use shared area for user/app")
	RootCmd.AddCommand(rmCmd)
}
