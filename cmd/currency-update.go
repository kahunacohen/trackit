/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var currencyUpdateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.ExactArgs(2),
	Short: "Updates an existing currency symbol. trackit currency update <old> <new>",
	Long:  `Updates an existing currency symbol. trackit currency update <old> <new>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("currencyUpdate called")
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyUpdateCmd)
}
