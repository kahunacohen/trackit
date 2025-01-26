/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "trackit",
	Short: "A CLI (command line application) for tracking personal finances",
	Long: `A CLI (command line application) for tracking personal finances. 
For example:

# initalizes your app by pointing by default to ~/.trackit.yaml
$ trackit init `,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

}
