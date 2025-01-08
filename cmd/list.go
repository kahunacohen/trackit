/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
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
		err = RenderTransactionTable(transactions)
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
	lsCmd.Flags().StringP("aggregate-by", "g", "", "What to aggregate by")
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

func RenderTransactionTable(rows []models.ReadTransactionsRow) error {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Date", "Payee", "Account", "Category", "Ignore", "Amount"})
	var total float64
	for _, row := range rows {
		var category string
		if row.CategoryName.Valid {
			category = row.CategoryName.String
		} else {
			category = "-"
		}
		ignoreVal := "No"
		if row.IgnoreWhenSumming == 1 {
			ignoreVal = "Yes"
		}
		t.AppendRow([]interface{}{row.TransactionID, row.Date, row.CounterParty, accountKeyToName(row.AccountName), category, ignoreVal, fmt.Sprintf("%.2f", row.Amount)})
		if row.IgnoreWhenSumming == 0 {
			total += row.Amount
		}
	}
	totalStr := strconv.FormatFloat(total, 'f', 2, 64) // 'f' for floating-point format, 2 digits after the decimal

	t.AppendFooter(table.Row{"", "", "", "", "", "Total", totalStr})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:  "Amount",
			Align: 4,
		},
	})
	t.Render()
	return nil
}

func accountKeyToName(account string) string {
	var name string
	split := strings.Split(account, "_")
	for i, s := range split {
		name += s
		if i != len(split)-1 {
			name += " "
		}
	}
	return strings.Title(name)
}

func getAccountTransactions(db *sql.DB, accountName string, date string) ([]models.ReadTransactionsRow, error) {
	var rows []models.ReadTransactionsRow
	var err error
	queries := models.New(db)
	// account and date are not set
	ctx := context.Background()

	// @TODO this is a bit messy, repeated code, etc. Maybe make a wrapper function
	// that handles the distinct types but with same fields.
	if accountName == "" && date == "" {
		rows, err = queries.ReadTransactions(ctx)
		if err != nil {
			return nil, err
		}
	} else if accountName != "" && date != "" {
		xs, err := queries.ReadTransactionsByAccountNameAndDate(ctx, models.ReadTransactionsByAccountNameAndDateParams{
			AccountName: accountName,
			Date:        date})
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
		// account name is set but not date
	} else if accountName != "" && date == "" {
		xs, err := queries.ReadTransactionsByAccountName(ctx, accountName)
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
		// date is set but not account
	} else {
		xs, err := queries.ReadTransactionsByDate(ctx, date)
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
	}
	if err != nil {
		return nil, err
	}
	return rows, nil
}
