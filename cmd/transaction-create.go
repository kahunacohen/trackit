/*
Copyright © 2025 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var transactionCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Short:   "creates a transaction",
	Long: `With create, you can create a transaction that is not listed in
one of your CSV files. For example, say somebody gives you cash as a gift.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		account, _ := flags.GetString("account")
		amount, _ := flags.GetFloat64("amount")
		categoryId, _ := flags.GetInt64("category-id")
		counterParty, _ := flags.GetString("counter-party")
		date, _ := flags.GetString("date")
		// description, _ := flags.GetString("description")
		ignore, _ := flags.GetBool("ignore")
		if amount != 0 && counterParty != "" && date != "" {
			ctx := context.Background()
			if !validateDateWithDayFormat(date) {
				return fmt.Errorf("date '%s' is invalid. Must be in form: YYYY/mm/dd", date)
			}
			db, err := getDB()
			if err != nil {
				return err
			}
			queries := models.New(db)
			var accountIdNullInt64 sql.NullInt64
			if account != "" {
				accountId, err := queries.ReadAccountIdByName(ctx, account)
				if err != nil {
					return fmt.Errorf("error getting account ID: %w", err)
				}
				accountIdNullInt64 = sql.NullInt64{Valid: true, Int64: accountId}
			} else {
				accountIdNullInt64 = sql.NullInt64{Valid: false}
			}
			err = queries.CreateTransaction(ctx, models.CreateTransactionParams{
				AccountID: accountIdNullInt64,
				Amount:    amount,
				CategoryID: func() sql.NullInt64 {
					return sql.NullInt64{Valid: categoryId != 0, Int64: categoryId}
				}(),
				CounterParty: counterParty,
				Date:         date,
				IgnoreWhenSumming: func() int64 {
					if ignore {
						return 1
					} else {
						return 0
					}
				}(),
			})
			if err != nil {
				return fmt.Errorf("error creating transaction: %w", err)
			}

		} else {
			return errors.New("must pass at least amount, counter-payer and date flags")
		}
		return nil

	},
}

func init() {
	transactionCreateCmd.Flags().StringP("account", "a", "", "account key")
	transactionCreateCmd.Flags().Float64P("amount", "m", 0, "amount of transaction")
	transactionCreateCmd.Flags().StringP("counter-party", "c", "", "other party participating in transaction")
	transactionCreateCmd.Flags().StringP("description", "d", "", "description of transaction")
	transactionCreateCmd.Flags().BoolP("ignore", "i", false, "whether to ignore amount when summing or aggregating")
	transactionCreateCmd.Flags().Int64P("category-id", "t", 0, "an existing category ID. Do trackit categories list to see existing categories.")
	transactionCreateCmd.Flags().StringP("date", "e", "", "Date in YYYY-MM-DD format")
	transactionCmd.AddCommand(transactionCreateCmd)
}
