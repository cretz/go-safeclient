package cmd

import (
	"errors"
	"github.com/cretz/go-safeclient/client"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var putShared bool
var putFromFile string
var putOffset int64

var putCmd = &cobra.Command{
	Use:   "put [file path]",
	Short: "Put file contents",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("One and only one argument allowed")
		}
		c, err := getClient()
		if err != nil {
			log.Fatalf("Unable to obtain client: %v", err)
		}
		input := os.Stdin
		if putFromFile != "" {
			input, err = os.Open(putFromFile)
			if err != nil {
				log.Fatalf("Unable to read file: %v", err)
			}
			defer input.Close()
		}
		info := client.WriteFileInfo{
			FilePath: args[0],
			Shared:   putShared,
			Contents: input,
			Offset:   putOffset,
		}
		if err = c.WriteFile(info); err != nil {
			log.Fatalf("Failed to write file: %v", err)
		}
		return nil
	},
}

func init() {
	putCmd.Flags().BoolVarP(&putShared, "shared", "s", false, "Use shared area for user/app")
	putCmd.Flags().StringVarP(&putFromFile, "file", "f", "", "Read from a file instead of stdin")
	putCmd.Flags().Int64VarP(&putOffset, "offset", "o", 0, "Offset to start writing from")
	RootCmd.AddCommand(putCmd)
}
