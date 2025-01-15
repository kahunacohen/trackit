/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches transactions for text. trackit search <text>",
	Long: `Searches the transaction counter payer for text. Can filter by account and date. E.g.
trackit search <text>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		// @TODO add this in filtering
		fmt.Println(date)
		fmt.Println(account)
		db, err := getDB()
		if err != nil {
			return err
		}
		queries := models.New(db)
		var transactions []models.TransactionsView
		ts, err := queries.SearchTransactionsWithSum(context.Background(), sql.NullString{Valid: true, String: args[0]})
		var total float64
		if len(ts) > 0 {
			total = ts[0].TotalAmount.Float64
		}
		if err != nil {
			return fmt.Errorf("error searching transactions: %w", err)
		}
		for _, t := range ts {
			transactions = append(transactions, models.TransactionsView{
				AccountID:         t.AccountID,
				AccountName:       t.AccountName,
				TransactionID:     t.TransactionID,
				Date:              t.Date,
				CounterParty:      t.CounterParty,
				Amount:            t.Amount,
				IgnoreWhenSumming: t.IgnoreWhenSumming,
				Description:       t.Description,
				CategoryName:      t.CategoryName,
			})
		}
		renderTransactionTable(transactions, &total)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format. For now, day precision is not implemented.")
	searchCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
}
