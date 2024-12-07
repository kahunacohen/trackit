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
			isValid, err := validateDateFormat(date)
			if err != nil {
				return err
			}
			if !isValid {
				return fmt.Errorf("date must be in MM-YYYY format")
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
		fmt.Println(aggregateBy)
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
				return fmt.Errorf("error rendering transactions")
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

func RenderTransactionTable(transactions []database.Transaction) error {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Date", "Payee", "Category", "Amount"})
	var total float64
	for _, transaction := range transactions {
		parsedTime, err := time.Parse(time.RFC3339, transaction.Date)
		if err != nil {
			return err
		}
		formattedDate := parsedTime.Format("01-02-2006")
		var cat string
		if transaction.Category == nil {
			cat = "-"
		} else {
			cat = *transaction.Category
		}
		t.AppendRow([]interface{}{formattedDate, transaction.CounterParty, cat, fmt.Sprintf("%.2f", transaction.Amount)})
		total += transaction.Amount
	}
	totalStr := strconv.FormatFloat(total, 'f', 2, 64) // 'f' for floating-point format, 2 digits after the decimal

	t.AppendFooter(table.Row{"", "", "Total", totalStr})
	t.Render()
	return nil
}
