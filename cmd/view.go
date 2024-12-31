/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/config"
	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

func validateDateFormat(date string) (bool, error) {
	// Regular expression for MM-YYYY format
	re := regexp.MustCompile(`^(0[1-9]|1[0-2])-[0-9]{4}$`)
	return re.MatchString(date), nil
}

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Generates a view of account data",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
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
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().StringP("date", "d", "", "Date in MM-YYYY format")
	viewCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
	viewCmd.Flags().StringP("aggregate-by", "g", "", "What to aggregate by")

	viewCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Get the date flag value
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		aggregateBy, _ := cmd.Flags().GetString("aggregate-by")
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		if aggregateBy == "" {
			transactions, err := database.GetAccountTransactions(db, account, date)
			if err != nil {
				return fmt.Errorf("error getting transactions: %w", err)
			}
			err = RenderTransactionTable(transactions)
			if err != nil {
				return fmt.Errorf("error rendering transactions: %w", err)
			}
		} else if aggregateBy == "category" {
			aggregations, err := database.GetCategoryAggregation(db, account, date)
			if err != nil {
				return fmt.Errorf("error aggregating by category: %w", err)
			}
			RenderAggregateTable(aggregations)
		} else {
			return fmt.Errorf("invalid aggregation '%s'", aggregateBy)
		}
		return nil
	}
}

func RenderAggregateTable(aggregates []database.CategoryAgregation) {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Category", "Total"})
	for _, aggregate := range aggregates {
		t.AppendRow([]interface{}{aggregate.Category, aggregate.Total})
	}
	t.Render()
}

func RenderTransactionTable(rows []models.ReadAllTransactionsRow) error {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Date", "Payee", "Account", "Category", "Amount"})
	var total float64
	for _, row := range rows {
		var category string
		if row.CategoryName.Valid {
			category = row.CategoryName.String
		} else {
			category = "-"
		}
		t.AppendRow([]interface{}{row.TransactionID, row.Date, row.CounterParty, row.AccountName, category, fmt.Sprintf("%.2f", row.Amount)})
		total += row.Amount
	}
	totalStr := strconv.FormatFloat(total, 'f', 2, 64) // 'f' for floating-point format, 2 digits after the decimal

	t.AppendFooter(table.Row{"", "", "", "", "Total", totalStr})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:  "Amount",
			Align: 4,
		},
	})
	t.Render()
	return nil
}
