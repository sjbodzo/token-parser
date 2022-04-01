package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "development" // note: this is replaced during 'go build'

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of coinparser",
	Long:  `Coinparser tagged build version`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Version: ", Version)
	},
}
