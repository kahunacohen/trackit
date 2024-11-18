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
		if date == "" && account == "" {
			return fmt.Errorf("either date or account flag is required")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().StringP("date", "d", "", "Date in MM-YYYY format")
	viewCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
	viewCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Get the date flag value
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")

		// Check if the date is in the correct MM-YYYY format using regex
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

		if account != "" && date == "" {
			fmt.Println("get account transactions")
			homeDir, _ := os.UserHomeDir()
			dbPath := filepath.Join(homeDir, "trackit.db")
			db, err := database.GetDB(dbPath)
			if err != nil {
				log.Fatalf("Failed to open database: %v", err)
			}
			transactions, err := database.GetAccountTransactions(db, account)
			if err != nil {
				return fmt.Errorf("error getting transactions: %w", err)
			}
			t := table.NewWriter()
			t.SetStyle(table.StyleLight)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Date", "Payee", "Amount", "Category"})
			for _, transaction := range transactions {
				var cat string
				if transaction.Category == nil {
					cat = "-"
				} else {
					cat = *transaction.Category
				}
				t.AppendRow([]interface{}{transaction.Date, transaction.CounterParty, transaction.Amount, cat})
			}
			t.Render()
		}

		return nil
	}
}
