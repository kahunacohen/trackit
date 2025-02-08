/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var currencyCmd = &cobra.Command{
	Use:   "currency",
	Short: "Manages currencies for currency conversions.",
	Long:  `Manages currencies for currency conversions.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(currencyCmd)
}
