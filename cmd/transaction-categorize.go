/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

const skipText = "Skip categorizing this transaction"

var categoryNames []string

var transactionCategorizeCmd = &cobra.Command{
	Use:     "categorize",
	Aliases: []string{"cat"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("too many arguments, expected at most 1")
		}
		return nil
	},
	Short: "Categorizes transactions",
	Long: `categorize transactions either by interactively categorizing all un-categorized
transactions (no flags passed), or by categorizing an individual transaction by ID (trackit categorize <transaction_id>).
Get the transaction ID by doing trackit list. To update existing transaction categories, just run trackit categorize (or)
trackit categorize <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var transactionId int
		var strconvErr error
		if len(args) == 1 {
			transactionId, strconvErr = strconv.Atoi(args[0])
			if strconvErr != nil {
				return fmt.Errorf("error parsing transaction id: %w", strconvErr)
			}
		}
		db, err := getDB()
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
		categoryNames = append(categoryNames, skipText)
		for _, category := range categories {
			categoryNames = append(categoryNames, category.Name)
		}
		if transactionId == 0 {
			transactions, err := queries.ReadNonCategorizedTransactions(ctx)
			if err != nil {
				return fmt.Errorf("error reading non categorized transactions: %w", err)
			}
			for _, transaction := range transactions {
				t := table.NewWriter()
				t.SetStyle(table.StyleLight)
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"Date", "Account", "Payee", "Amount"})
				var accountKey string
				if transaction.AccountName.Valid {
					accountKey = transaction.AccountName.String
				} else {
					return errors.New("account is null")
				}

				t.AppendRow([]interface{}{transaction.Date,
					accountKeyToName(sql.NullString{Valid: true, String: accountKey}),
					transaction.CounterParty, fmt.Sprintf("%.2f", transaction.Amount)})

				prompt := promptui.Select{
					Label:             t.Render(),
					Items:             categoryNames,
					StartInSearchMode: true,
					Searcher:          seacher,
				}
				_, categoryNameResult, err := prompt.Run()
				if err != nil {
					return fmt.Errorf("prompt failed %w", err)
				}
				if categoryNameResult == skipText {
					continue
				}
				err = queries.UpdateTransactionCategory(ctx, models.UpdateTransactionCategoryParams{
					CategoryID: sql.NullInt64{Valid: true, Int64: categoryMap[categoryNameResult]},
					ID:         transaction.TransactionID})
				if err != nil {
					return fmt.Errorf("error setting category: %w", err)
				}
			}

		} else {
			transaction, err := queries.ReadTransactionById(ctx, int64(transactionId))
			if err != nil {
				return fmt.Errorf("error getting transaction %d", transactionId)
			}
			t := table.NewWriter()
			t.SetStyle(table.StyleLight)
			t.SetOutputMirror(os.Stdout)
			t.AppendHeader(table.Row{"Date", "Account", "Payee", "Amount"})
			t.AppendRow([]interface{}{transaction.Date, transaction.AccountName.String, transaction.CounterParty, fmt.Sprintf("%.2f", transaction.Amount)})
			prompt := promptui.Select{
				Label:    t.Render(),
				Items:    categoryNames,
				Searcher: seacher,
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
	transactionCmd.AddCommand(transactionCategorizeCmd)
}

func seacher(input string, i int) bool {
	return strings.Contains(strings.ToLower(categoryNames[i]), strings.ToLower(input))
}
