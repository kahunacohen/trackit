/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var transactionListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Lists transactions",
	Long:    `Lists transactions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		_, _, dbPath, err := getDataPaths()
		if err != nil {
			return err
		}
		db, err := getDB(dbPath)
		if err != nil {
			return err
		}
		if date != "" {
			_, err := time.Parse("2006-01", date)
			if err != nil {
				return fmt.Errorf("date must be in YYYY-MM format")
			}
		}
		if account != "" {
			queries := models.New(db)
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
		transactions, total, err := getAccountTransactions(db, account, date)
		if err != nil {
			return fmt.Errorf("error getting transactions: %w", err)
		}

		err = renderTransactionTable(transactions, total)
		if err != nil {
			return fmt.Errorf("error rendering transactions: %w", err)
		}
		return nil
	},
}

func init() {
	transactionCmd.AddCommand(transactionListCmd)
	transactionListCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format. For now, day precision is not implemented.")
	transactionListCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
}

func RenderAggregateTable(aggregates []models.AggregateTransactionsRow) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Category", "Total"})
	for _, aggregate := range aggregates {
		t.AppendRow([]interface{}{aggregate.CategoryName, fmt.Sprintf("%.2f", aggregate.TotalAmount)})
	}
	t.Render()
}

func getAccountTransactions(db *sql.DB, accountName string, date string) ([]models.TransactionsView, *float64, error) {
	var transactions []models.TransactionsView
	var total float64
	queries := models.New(db)
	// account and date are not set
	ctx := context.Background()

	// @TODO this is a bit messy, repeated code, etc. Maybe make a wrapper function
	// that handles the distinct types but with same fields.
	if accountName == "" && date == "" {
		ts, err := queries.ReadTransactionsWithSum(ctx)
		if err != nil {
			return nil, nil, err
		}
		if len(ts) > 0 {
			total = ts[0].TotalAmount.Float64
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
	} else if accountName != "" && date != "" {
		ts, err := queries.ReadTransactionsByAccountNameAndDateWithSum(ctx, models.ReadTransactionsByAccountNameAndDateWithSumParams{
			AccountName: sql.NullString{Valid: true, String: accountName},
			Date:        date})
		if err != nil {
			return nil, nil, err
		}
		if len(ts) > 0 {
			total = ts[0].TotalAmount.Float64
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

		// account name is set but not date
	} else if accountName != "" && date == "" {
		ts, err := queries.ReadTransactionsByAccountNameWithSum(ctx, sql.NullString{Valid: true, String: accountName})
		if err != nil {
			return nil, nil, err
		}
		if len(ts) > 0 {
			total = ts[0].TotalAmount.Float64
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
		ts, err := queries.ReadTransactionsByDateWithSum(ctx, date)
		if err != nil {
			return nil, nil, err
		}
		if len(ts) > 0 {
			total = ts[0].TotalAmount.Float64
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
	return transactions, &total, nil
}
