/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// currencyCreateCmd represents the currencyCreate command
var currencyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates a new ISO 4217 currency symbol. trackit currency create ILS",
	Long:  `Creates a new ISO 4217 three-letter ISO symbol. "USD" is pre-populated. trackit currency create ILS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyCreateCmd)
}
