package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"log"
)

var lsShared bool

var lsCmd = &cobra.Command{
	Use:   "ls [dir]",
	Short: "Fetch directory information",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalf("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		err = c.GetDir(args[0], lsShared)
		if err != nil {
			log.Fatalf("NO: %v", err)
		}
		fmt.Println("Ping successful")
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsShared, "shared", "s", false, "Use shared area for user")
	RootCmd.AddCommand(lsCmd)
}
