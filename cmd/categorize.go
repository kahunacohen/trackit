/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// categorizeCmd represents the categorize command
var categorizeCmd = &cobra.Command{
	Use:   "categorize",
	Short: "Categorizes transactions",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		if interactive {
			queries := models.New(db)
			rows, err := queries.ReadNonCategorizedTransactions(ctx)
			if err != nil {
				return fmt.Errorf("error reading non categorized transactions: %w", err)
			}
			categories, err := queries.ReadAllCategories(ctx)
			if err != nil {
				return fmt.Errorf("error getting categories: %w", err)
			}

			var categoryMap map[string]int64 = make(map[string]int64)
			for _, category := range categories {
				categoryMap[category.Name] = category.ID
			}
			var categoryNames []string
			for categoryName, _ := range categoryMap {
				categoryNames = append(categoryNames, categoryName)
			}

			for _, row := range rows {
				prompt := promptui.Select{
					Label: fmt.Sprintf("Select a category for account %s, to %s for %.2f on %s",
						row.AccountName, row.CounterParty, row.Amount, row.Date.Format("01-02-2006")),
					Items: categoryNames,
				}
				_, categoryNameResult, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("prompt failed %w", err)
				}
				err = queries.UpdateTransactionCategory(ctx, models.UpdateTransactionCategoryParams{
					CategoryID: sql.NullInt64{Valid: true, Int64: categoryMap[categoryNameResult]},
					ID:         categoryMap[categoryNameResult]})
				if err != nil {
					return fmt.Errorf("error setting category: %w", err)
				}
			}

		}
		return nil
	},
}
var interactive bool

func init() {
	categorizeCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive mode")
	rootCmd.AddCommand(categorizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// categorizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// categorizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
