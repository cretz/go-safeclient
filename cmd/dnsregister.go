package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
)

var dnsRegisterShared bool

var dnsRegisterCmd = &cobra.Command{
	Use:   "dnsregister [name] [service name] [home dir path]",
	Short: "Register DNS name, service, and path",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 3 {
			return errors.New("Must have exactly 3 arguments for name, service name, and home dir path")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to get client: %v", err)
		}
		info := client.DNSRegisterInfo{
			Name:        args[0],
			ServiceName: args[1],
			HomeDirPath: args[2],
			Shared:      dnsRegisterShared,
		}
		if err = c.DNSRegister(info); err != nil {
			log.Fatalf("Unable to register: %v", err)
		}
		return nil
	},
}

func init() {
	dnsRegisterCmd.Flags().BoolVarP(&dnsRegisterShared, "shared", "s", false, "Use shared area for user/app")
	RootCmd.AddCommand(dnsRegisterCmd)
}
