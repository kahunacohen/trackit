/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

// lsCmd represents the view command
var lsCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists transactions",
	Long:  `ls lists transactions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		db, err := getDB()
		if err != nil {
			return err
		}
		transactions, err := getAccountTransactions(db, account, date)
		if err != nil {
			return fmt.Errorf("error getting transactions: %w", err)
		}
		err = renderTransactionTable(transactions)
		if err != nil {
			return fmt.Errorf("error rendering transactions: %w", err)
		}
		if date != "" {
			_, err := time.Parse("2006-01", date)
			if err != nil {
				return fmt.Errorf("date must be in YYYY-MM format")
			}
		}
		if account != "" {
			conf, err := config.ParseConfig("./trackit.yaml")
			if err != nil {
				return fmt.Errorf("error parsing config: %v", err)
			}
			_, ok := conf.Accounts[account]
			if !ok {
				return fmt.Errorf("invalid account specified: %s. Check your config for valid account keys", account)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
	lsCmd.Flags().StringP("date", "d", "", "Date in YYYY-MM format")
	lsCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
}

func RenderAggregateTable(aggregates []models.AggregateTransactionsRow) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Category", "Total"})
	for _, aggregate := range aggregates {
		if aggregate.TotalAmount.Valid {
			t.AppendRow([]interface{}{aggregate.CategoryName, roundAmount(aggregate.TotalAmount.Float64)})
		}
	}
	t.Render()
}

func accountKeyToName(account sql.NullString) string {
	if !account.Valid {
		return "-"
	}
	var name string
	split := strings.Split(account.String, "_")
	for i, s := range split {
		name += s
		if i != len(split)-1 {
			name += " "
		}
	}
	return strings.Title(name)
}

func getAccountTransactions(db *sql.DB, accountName string, date string) ([]models.TransactionsView, error) {
	var transactions []models.TransactionsView
	var err error
	queries := models.New(db)
	// account and date are not set
	ctx := context.Background()

	// @TODO this is a bit messy, repeated code, etc. Maybe make a wrapper function
	// that handles the distinct types but with same fields.
	if accountName == "" && date == "" {
		transactions, err = queries.ReadTransactions(ctx)
		if err != nil {
			return nil, err
		}
	} else if accountName != "" && date != "" {
		transactions, err = queries.ReadTransactionsByAccountNameAndDate(ctx, models.ReadTransactionsByAccountNameAndDateParams{
			AccountName: sql.NullString{Valid: true, String: accountName},
			Date:        date})
		if err != nil {
			return nil, err
		}

		// account name is set but not date
	} else if accountName != "" && date == "" {
		transactions, err = queries.ReadTransactionsByAccountName(ctx, sql.NullString{Valid: true, String: accountName})
		if err != nil {
			return nil, err
		}
		// date is set but not account
	} else {
		transactions, err = queries.ReadTransactionsByDate(ctx, date)
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
	}
	return transactions, nil
}
