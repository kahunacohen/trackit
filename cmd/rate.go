/*
Copyright © 2025 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// rateCmd represents the rate command
var rateCmd = &cobra.Command{
	Use:   "rate",
	Short: "Manages rates for currency conversion.",
	Long:  `Manages rates for currency conversion.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(rateCmd)
}
