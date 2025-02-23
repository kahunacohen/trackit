/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/kahunacohen/trackit/internal/models"

	"github.com/spf13/cobra"
)

var transactionAggregateCmd = &cobra.Command{
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
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		if by == "category" {
			aggregations, err := getCategoryAggregation(db, account, date)
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
	transactionAggregateCmd.Flags().StringP("account", "a", "", "account key from trackit.yaml to filter by account")
	transactionAggregateCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format. For now, day precision is not implemented.")
	transactionAggregateCmd.Flags().StringP("by", "b", "category", "What to aggregate total by")
	transactionCmd.AddCommand(transactionAggregateCmd)
}

func getCategoryAggregation(db *sql.DB, account string, date string) ([]models.AggregateTransactionsRow, error) {
	queries := models.New(db)
	ctx := context.Background()
	var err error
	var rows []models.AggregateTransactionsRow
	if account == "" && date == "" {
		rows, err = queries.AggregateTransactions(ctx)
		if err != nil {
			return nil, fmt.Errorf("error aggregating rows: %w", err)
		}
	} else if account != "" && date == "" {
		xs, err := queries.AggregateTransactionsByAccountName(ctx, sql.NullString{Valid: true, String: account})
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	} else if account == "" && date != "" {
		xs, err := queries.AggregateTransactionsByDate(ctx, date)
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	} else {
		xs, err := queries.AggregateTransactionsByAccountNameAndDate(ctx,
			models.AggregateTransactionsByAccountNameAndDateParams{AccountName: sql.NullString{Valid: true, String: account}, Date: date})
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	}
	return rows, nil
}
