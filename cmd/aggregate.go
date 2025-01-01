/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// aggregateCmd represents the aggregate command
var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregates transactions by some facet, reporting total amount (default is category)",
	Long: `Aggregates transactions by some facet (default is category): E.g.
	
$ trackit aggregate
$ trackit aggregate --by account
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("aggregate called")
	},
}

func init() {
	aggregateCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format")
	aggregateCmd.Flags().StringP("by", "b", "category", "What to aggregate total by")
	rootCmd.AddCommand(aggregateCmd)
}
