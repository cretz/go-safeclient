package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var mvShared bool
var mvDestChangeShared bool

var mvCmd = &cobra.Command{
	Use: "mv [src file] [dest dir]",
	// TODO: https://maidsafe.atlassian.net/browse/CS-60
	Short: "Move file [NOT YET WORKING]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Exactly two arguments required for source and destination")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		destShared := mvShared
		if mvDestChangeShared {
			destShared = !mvShared
		}
		info := client.MoveFileInfo{
			SrcPath:      args[0],
			SrcShared:    mvShared,
			DestPath:     args[1],
			DestShared:   destShared,
			RetainSource: false,
		}
		if err = c.MoveFile(info); err != nil {
			log.Fatalf("Failed to move file: %v", err)
		}
		return nil
	},
}

func init() {
	mvCmd.Flags().BoolVarP(&mvShared, "shared", "s", false, "Use shared area for user/app")
	mvCmd.Flags().BoolVar(&mvDestChangeShared, "change-shared", false, "Change whether the destination is shared based on the source")
	RootCmd.AddCommand(mvCmd)
}
