package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var dnsAddServiceShared bool

var dnsAddServiceCmd = &cobra.Command{
	Use:   "dnsaddservice [name] [service name] [home dir path]",
	Short: "Add DNS service and path to existing DNS name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return errors.New("Must have exactly 3 arguments for name, service name, and home dir path")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		info := client.DNSAddServiceInfo{
			Name:        args[0],
			ServiceName: args[1],
			HomeDirPath: args[2],
			Shared:      dnsRegisterShared,
		}
		if err = c.DNSAddService(info); err != nil {
			log.Fatalf("Unable to add service: %v", err)
		}
		return nil
	},
}

func init() {
	dnsAddServiceCmd.Flags().BoolVarP(&dnsAddServiceShared, "shared", "s", false, "Use shared area for user/app")
	RootCmd.AddCommand(dnsAddServiceCmd)
}
