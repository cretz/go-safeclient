package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var lsShared bool

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "Fetch directory information",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		info := client.GetDirInfo{DirPath: args[0], Shared: lsShared}
		dir, err := c.GetDir(info)
		if err != nil {
			log.Fatalf("Failed to list dir: %v", err)
		}
		writeDirResponseTable(os.Stdout, dir)
		return nil
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsShared, "shared", "s", false, "Use shared area for user")
	RootCmd.AddCommand(lsCmd)
}
