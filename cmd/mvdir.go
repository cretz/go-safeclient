package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var mvdirShared bool
var mvdirDestChangeShared bool

var mvdirCmd = &cobra.Command{
	Use:   "mvdir [src dir] [dest base dir]",
	Short: "Move directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Exactly two arguments required for source and destination")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		destShared := mvdirShared
		if mvdirDestChangeShared {
			destShared = !mvdirShared
		}
		info := client.MoveDirInfo{
			SrcPath:      args[0],
			SrcShared:    mvdirShared,
			DestPath:     args[1],
			DestShared:   destShared,
			RetainSource: false,
		}
		if err = c.MoveDir(info); err != nil {
			log.Fatalf("Failed to move dir: %v", err)
		}
		return nil
	},
}

func init() {
	mvdirCmd.Flags().BoolVarP(&mvdirShared, "shared", "s", false, "Use shared area for user/app")
	mvdirCmd.Flags().BoolVar(&mvdirDestChangeShared, "change-shared", false, "Change whether the destination is shared based on the source")
	RootCmd.AddCommand(mvdirCmd)
}
