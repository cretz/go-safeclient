package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var mkdirPrivate bool
var mkdirShared bool
var mkdirVersioned bool
var mkdirMetadata string

var mkdirCmd = &cobra.Command{
	Use:   "mkdir [dir]",
	Short: "Create directory",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.CreateDirInfo{
			DirPath:   args[0],
			Private:   mkdirPrivate,
			Versioned: mkdirVersioned,
			Metadata:  mkdirMetadata,
			Shared:    mkdirShared,
		}
		if err = c.CreateDir(info); err != nil {
			log.Fatalf("Failed to create dir: %v", err)
		}
		return nil
	},
}

func init() {
	mkdirCmd.Flags().BoolVarP(&mkdirPrivate, "private", "p", false, "Whether the directory is private")
	mkdirCmd.Flags().BoolVarP(&mkdirShared, "shared", "s", false, "Use shared area for user/app")
	mkdirCmd.Flags().BoolVar(&mkdirVersioned, "versioned", false, "Whether directory is versioned")
	mkdirCmd.Flags().StringVarP(&mkdirMetadata, "metadata", "m", "", "Metadata to store with the directory")
	RootCmd.AddCommand(mkdirCmd)
}
