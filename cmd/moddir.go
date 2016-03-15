package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"errors"
	"github.com/cretz/go-safeclient/client"
)

var moddirShared bool
var moddirName string
var moddirMetadata string

var moddirCmd = &cobra.Command{
	Use:   "moddir [dir]",
	Short: "Change directory name and/or metadata",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		} else if (moddirName == "" && moddirMetadata == "") {
			return errors.New("Must provide at least name or metadata")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.ChangeDirInfo{
			DirPath: args[0],
			Shared: moddirShared,
			NewName: moddirName,
			Metadata: moddirMetadata,
		}
		if err = c.ChangeDir(info); err != nil {
			log.Fatalf("Failed to change dir: %v", err)
		}
		return nil
	},
}

func init() {
	moddirCmd.Flags().BoolVarP(&moddirShared, "shared", "s", false, "Use shared area for user/app")
	moddirCmd.Flags().StringVarP(&moddirName, "name", "n", "", "The name to change to")
	moddirCmd.Flags().StringVarP(&moddirMetadata, "metadata", "m", "", "Metadata to update")
	RootCmd.AddCommand(moddirCmd)
}
