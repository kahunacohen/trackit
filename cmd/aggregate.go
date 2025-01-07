/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	database "github.com/kahunacohen/trackit/internal/db"

	"github.com/spf13/cobra"
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Aggregates transactions by some facet, reporting total amount (default is category)",
	Long: `Aggregates transactions by some facet (default is category): E.g.
	
$ trackit aggregate
$ trackit aggregate --by account
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		by, _ := cmd.Flags().GetString("by")
		db, err := database.GetDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		if by == "category" {
			aggregations, err := database.GetCategoryAggregation(db, account, date)
			if err != nil {
				return fmt.Errorf("error aggregating by category: %w", err)
			}
			RenderAggregateTable(aggregations)
		} else {
			return fmt.Errorf("aggregation '%s' not implemented yet", by)
		}
		return nil

	},
}

func init() {
	aggregateCmd.Flags().StringP("account", "a", "", "account key from trackit.yaml to filter by account")
	aggregateCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format with which to filter")
	aggregateCmd.Flags().StringP("by", "b", "category", "What to aggregate total by")
	rootCmd.AddCommand(aggregateCmd)
}
