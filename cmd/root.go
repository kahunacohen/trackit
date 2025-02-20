/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "trackit",
	Short:   "A CLI (command line application) for tracking personal finances",
	Version: "0.1.0",
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
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose mode")
}

func logLn(msg string, verbose bool) {
	if verbose {
		log.Println(msg)
	}
}
func logF(verbose bool, msg string, args ...interface{}) {
	if verbose {
		log.Printf(msg, args...)
	}
}
