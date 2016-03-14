package cmd

import (
	"github.com/spf13/cobra"
)

var dnsCmd = &cobra.Command{
	Use:   "dns",
	Short: "Do things with DNS",
}

func init() {
	RootCmd.AddCommand(dnsCmd)
}

