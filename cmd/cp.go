package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var cpShared bool
var cpDestChangeShared bool

var cpCmd = &cobra.Command{
	Use:   "cp [src file] [dest dir]",
	Short: "Copy file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return errors.New("Exactly two arguments required for source and destination")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		destShared := mvShared
		if cpDestChangeShared {
			destShared = !cpShared
		}
		info := client.MoveFileInfo{
			SrcPath:      args[0],
			SrcShared:    cpShared,
			DestPath:     args[1],
			DestShared:   destShared,
			RetainSource: true,
		}
		if err = c.MoveFile(info); err != nil {
			log.Fatalf("Failed to copy file: %v", err)
		}
		return nil
	},
}

func init() {
	cpCmd.Flags().BoolVarP(&cpShared, "shared", "s", false, "Use shared area for user/app")
	cpCmd.Flags().BoolVar(&cpDestChangeShared, "change-shared", false, "Change whether the destination is shared based on the source")
	RootCmd.AddCommand(cpCmd)
}
