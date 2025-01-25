/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var rateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "creates a rate for a given month",
	Long:  `creates a rate for a given month`,
	RunE: func(cmd *cobra.Command, args []string) error {
		monthParam, _ := cmd.Flags().GetString("month")
		if !validateYearMonthFormat(monthParam) {
			return fmt.Errorf("month param  \"%s\" is not valid month format (YYYY-MM)", monthParam)
		}
		fromSymbol, _ := cmd.Flags().GetString("fromSymbol")
		toSymbol, _ := cmd.Flags().GetString("toSymbol")
		fmt.Println(fromSymbol)
		fmt.Println(toSymbol)
		return nil
	},
}

func init() {
	rateCreateCmd.Flags().StringP("month", "m", "", "month in YYYY-MM format for the rate")
	rateCreateCmd.MarkFlagRequired("month")
	rateCreateCmd.Flags().StringP("from-symbol", "f", "", "from currency symbol")
	rateCreateCmd.MarkFlagRequired("from-symbol")
	rateCreateCmd.Flags().StringP("to-symbol", "t", "", "to currency symbol")
	rateCreateCmd.MarkFlagRequired("to-symbol")
	rateCreateCmd.Flags().Float64P("rate", "r", 1, "rate")
	rateCreateCmd.MarkFlagRequired("rate")
	rateCmd.AddCommand(rateCreateCmd)
}
