/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var rateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "creates a rate for a given month",
	Long:  `creates a rate for a given month`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	rateCreateCmd.Flags().StringP("month", "m", "", "month in YYYY-MM format for the rate")
	rateCreateCmd.Flags().StringP("from-symbol", "f", "", "from currency symbol")
	rateCreateCmd.Flags().StringP("to-symbol", "t", "", "to currency symbol")
	rateCreateCmd.Flags().Float64P("rate", "r", 1, "rate")
	rateCmd.AddCommand(rateCreateCmd)
}
