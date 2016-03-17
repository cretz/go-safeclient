package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var cpdirShared bool
var cpdirDestChangeShared bool

var cpdirCmd = &cobra.Command{
	Use: "cpdir [src dir] [dest dir]",
	// TODO: https://maidsafe.atlassian.net/browse/CS-60
	Short: "Copy directory [NOT YET WORKING]",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Exactly two arguments required for source and destination")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		destShared := cpdirShared
		if cpdirDestChangeShared {
			destShared = !cpdirShared
		}
		info := client.MoveDirInfo{
			SrcPath:      args[0],
			SrcShared:    cpdirShared,
			DestPath:     args[1],
			DestShared:   destShared,
			RetainSource: true,
		}
		if err = c.MoveDir(info); err != nil {
			log.Fatalf("Failed to copy dir: %v", err)
		}
		return nil
	},
}

func init() {
	cpdirCmd.Flags().BoolVarP(&cpdirShared, "shared", "s", false, "Use shared area for user/app")
	cpdirCmd.Flags().BoolVar(&cpdirDestChangeShared, "change-shared", false, "Change whether the destination is shared based on the source")
	RootCmd.AddCommand(cpdirCmd)
}
