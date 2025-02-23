/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var transactionSearchCmd = &cobra.Command{
	Use:   "search",
	Short: "Searches transactions' counter_party field for text. trackit search <text>",
	Long: `Searches the transaction counter payer for text. Can filter by account and date. E.g.
trackit search <text>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		db, _, err := getDB()
		if err != nil {
			return err
		}
		queries := models.New(db)
		if date != "" {
			if date != "" {
				_, err := time.Parse("2006-01", date)
				if err != nil {
					return fmt.Errorf("date must be in YYYY-MM format")
				}
			}
		}
		if account != "" {
			configPath, err := queries.ReadSettingByName(context.Background(), "config-file")
			if err != nil {
				return fmt.Errorf("error getting config-file path from db: %w", err)
			}
			conf, err := config.ParseConfig(configPath)
			if err != nil {
				return fmt.Errorf("error parsing config: %v", err)
			}
			_, ok := conf.Accounts[account]
			if !ok {
				return fmt.Errorf("invalid account specified: %s. Check your config for valid account keys", account)
			}
		}
		var transactions []models.TransactionsView
		var total float64

		if date == "" && account == "" {
			ts, err := queries.SearchTransactionsWithSum(context.Background(), sql.NullString{Valid: true, String: args[0]})

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
		} else if date != "" && account == "" {
			ts, err := queries.SearchTransactionsByDateWithSum(context.Background(), models.SearchTransactionsByDateWithSumParams{
				SearchTerm: sql.NullString{Valid: true, String: args[0]},
				Date:       date,
			})

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
		} else {
			if date != "" {
				_, err := time.Parse("2006-01", date)
				if err != nil {
					return fmt.Errorf("date must be in YYYY-MM format")
				}
			}
			ts, err := queries.SearchTransactionsByAccountNameAndDateWithSum(context.Background(), models.SearchTransactionsByAccountNameAndDateWithSumParams{
				SearchTerm:  sql.NullString{Valid: true, String: args[0]},
				AccountName: sql.NullString{Valid: true, String: account},
				Date:        date,
			})

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
		}
		renderTransactionTable(transactions, &total)
		return nil
	},
}

func init() {
	transactionCmd.AddCommand(transactionSearchCmd)
	transactionSearchCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format. For now, day precision is not implemented.")
	transactionSearchCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
}
