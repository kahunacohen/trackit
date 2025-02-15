/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// transactionCmd represents the transaction command
var transactionCmd = &cobra.Command{
	Use:     "transaction",
	Aliases: []string{"tr"},
	Short:   "Manages transactions",
	Long:    `Manages transactions including creating, updating, and deleting.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Usage()
	},
}

func init() {
	rootCmd.AddCommand(transactionCmd)
}
