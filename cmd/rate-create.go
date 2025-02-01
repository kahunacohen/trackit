/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var rateCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "creates a rate for a given month",
	Long:  `creates a rate for a given month`,
	RunE: func(cmd *cobra.Command, args []string) error {
		month, _ := cmd.Flags().GetString("month")
		if !validateYearMonthFormat(month) {
			return fmt.Errorf("month param  \"%s\" is not valid month format (YYYY-MM)", month)
		}
		fromSymbol, _ := cmd.Flags().GetString("from-symbol")
		rate, _ := cmd.Flags().GetFloat64("rate")
		fromSymbol = strings.ToUpper(fromSymbol)
		db, err := getDB()
		if err != nil {
			return err
		}
		ctx := context.Background()
		queries := models.New(db)
		currencies, err := queries.ReadAllCurrencies(ctx)
		if err != nil {
			return fmt.Errorf("error reading currencies: %w", err)
		}
		var fromSymbolFound bool
		var fromCurrencyID int64
		for _, curr := range currencies {
			if fromSymbol == curr.Symbol {
				fromSymbolFound = true
				fromCurrencyID = curr.ID
			}
		}
		if !fromSymbolFound {
			return fmt.Errorf("fromSymbol \"%s\" not found. You must create it first with trackit currency create", fromSymbol)
		}
		if !validateYearMonthFormat(month) {
			return fmt.Errorf("month param \"%s\" must be in form YYYY-MM", month)
		}
		err = queries.CreateRate(ctx, models.CreateRateParams{
			Rate:               rate,
			CurrencyCodeFromID: fromCurrencyID,
			Month:              month})
		if err != nil {
			return fmt.Errorf("error creating rate: %w", err)
		}
		return nil
	},
}

func init() {
	rateCreateCmd.Flags().StringP("month", "m", "", "month in YYYY-MM format for the rate")
	rateCreateCmd.MarkFlagRequired("month")
	rateCreateCmd.Flags().StringP("from-symbol", "f", "", "from currency symbol")
	rateCreateCmd.MarkFlagRequired("from-symbol")
	rateCreateCmd.Flags().Float64P("rate", "r", 1, "rate")
	rateCreateCmd.MarkFlagRequired("rate")
	rateCmd.AddCommand(rateCreateCmd)
}
