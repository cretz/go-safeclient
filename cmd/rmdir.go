package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var rmdirShared bool

var rmdirCmd = &cobra.Command{
	Use:   "rmdir [dir]",
	Short: "Delete directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.DeleteDirInfo{
			DirPath: args[0],
			Shared:  rmdirShared,
		}
		if err = c.DeleteDir(info); err != nil {
			log.Fatalf("Failed to delete dir: %v", err)
		}
		return nil
	},
}

func init() {
	rmdirCmd.Flags().BoolVarP(&rmdirShared, "shared", "s", false, "Use shared area for user/app")
	RootCmd.AddCommand(rmdirCmd)
}
