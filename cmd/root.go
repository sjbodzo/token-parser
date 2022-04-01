package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "coinparser",
	Short: "Coin Parser is a crypto coin verifier",
	Long:  `An asynchronous, fast verification engine for parsing crypto coins based off of their IDs. `,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
