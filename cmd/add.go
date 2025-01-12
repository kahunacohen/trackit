/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
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

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds a transaction",
	Long: `With add, you can add a transaction that is not listed in
one of your CSV files. For example, say somebody gives you cash as a gift.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		amount, _ := flags.GetFloat64("amount")
		counterParty, _ := flags.GetString("counter-party")
		description, _ := flags.GetString("description")
		ignore, _ := flags.GetBool("ignore")
		category, _ := flags.GetString("category-id")
		date, _ := flags.GetString("date")
		if amount != 0 && counterParty != "" && date != "" {
			fmt.Println(description)
			fmt.Println(ignore)
			fmt.Println(category)
			if !validateDateWithDayFormat(date) {
				return fmt.Errorf("date '%s' is invalid. Must be in form: YYYY/mm/dd", date)
			}
			db, err := getDB()
			if err != nil {
				return err
			}
			queries := models.New(db)
			err = queries.CreateTransaction(context.Background(), models.CreateTransactionParams{
				AccountID:    sql.NullInt64{Valid: false},
				Date:         date,
				Amount:       amount,
				CounterParty: counterParty,
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
	addCmd.Flags().Float64P("amount", "a", 0, "amount")
	addCmd.Flags().StringP("counter-party", "c", "", "other party participating in transaction")
	addCmd.Flags().StringP("description", "d", "", "description of transaction")
	addCmd.Flags().BoolP("ignore", "i", false, "whether to ignore amount when summing or aggregating")
	addCmd.Flags().Int64P("category-id", "t", 0, "an existing category ID. Do trackit categories list to see existing categories.")
	addCmd.Flags().StringP("date", "e", "", "The date in the form YYYY/mm")
	rootCmd.AddCommand(addCmd)
}
