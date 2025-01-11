/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a transaction",
	Long: `With add, you can add a transaction that is not listed in
one of your CSV files. For example, say somebody gives you cash as a gift.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("add called")
	},
}

func init() {
	addCmd.Flags().Int64P("amount", "a", 0, "amount")
	addCmd.Flags().StringP("counter_payer", "c", "", "account key from trackit.yaml to filter by account")
	addCmd.Flags().StringP("description", "d", "", "description of transaction")
	addCmd.Flags().BoolP("ignore", "i", false, "whether to ignore amount when summing or aggregating")
	addCmd.Flags().StringP("category", "t", "", "an existing category. If it doesn't exist you must create it first")
	addCmd.Flags().StringP("date", "e", "", "The date in the form YYYY/mm")
	rootCmd.AddCommand(addCmd)
}
