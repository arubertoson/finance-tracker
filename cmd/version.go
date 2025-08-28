package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of Finance Tracker",
	Long:  `All software has versions. This is Finance Tracker's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Finance Tracker CLI version %s\n", Version)
	},
}