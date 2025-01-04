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

	"github.com/jedib0t/go-pretty/v6/table"
	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var categorizeCmd = &cobra.Command{
	Use:   "categorize",
	Short: "Categorizes transactions",
	Long: `categorize categorizes transactions either by interactively categorizing all un-categorized
transactions (no flags passed), or by categorizing individual transactions by ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		id, _ := cmd.Flags().GetInt64("	id")
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)

		categories, err := queries.ReadAllCategories(ctx)
		if err != nil {
			return fmt.Errorf("error getting categories: %w", err)
		}

		var categoryMap map[string]int64 = make(map[string]int64)
		for _, category := range categories {
			categoryMap[category.Name] = category.ID
		}
		var categoryNames []string
		for _, category := range categories {
			categoryNames = append(categoryNames, category.Name)
		}
		if id == 0 {
			transactions, err := queries.ReadNonCategorizedTransactions(ctx)
			if err != nil {
				return fmt.Errorf("error reading non categorized transactions: %w", err)
			}
			for _, transaction := range transactions {
				t := table.NewWriter()
				t.SetStyle(table.StyleLight)
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"Date", "Account", "Payee", "Amount"})
				t.AppendRow([]interface{}{transaction.Date, transaction.AccountName, transaction.CounterParty, fmt.Sprintf("%.2f", transaction.Amount)})
				prompt := promptui.Select{
					Label: t.Render(),
					Items: categoryNames,
				}
				_, categoryNameResult, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("prompt failed %w", err)
				}
				err = queries.UpdateTransactionCategory(ctx, models.UpdateTransactionCategoryParams{
					CategoryID: sql.NullInt64{Valid: true, Int64: categoryMap[categoryNameResult]},
					ID:         transaction.ID})
				if err != nil {
					return fmt.Errorf("error setting category: %w", err)
				}
			}

		} else {
			transaction, err := queries.ReadTransactionById(ctx, id)
			if err != nil {
				return fmt.Errorf("error getting transaction %d", id)
			}
			t := table.NewWriter()
			t.SetStyle(table.StyleLight)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Date", "Account", "Payee", "Amount"})
			t.AppendRow([]interface{}{transaction.Date, transaction.AccountName, transaction.CounterParty, fmt.Sprintf("%.2f", transaction.Amount)})
			prompt := promptui.Select{
				Label: t.Render(),
				Items: categoryNames,
			}
			_, categoryNameResult, err := prompt.Run()
			if err != nil {
				return fmt.Errorf("prompt failed %w", err)
			}
			err = queries.UpdateTransactionCategory(ctx, models.UpdateTransactionCategoryParams{
				CategoryID: sql.NullInt64{Valid: true, Int64: categoryMap[categoryNameResult]},
				ID:         transaction.TransactionID})
			if err != nil {
				return fmt.Errorf("error setting category: %w", err)
			}
		}
		return nil
	},
}

func init() {
	categorizeCmd.Flags().Int64P("id", "i", 0, "valid transaction ID to categorize")
	rootCmd.AddCommand(categorizeCmd)
}
