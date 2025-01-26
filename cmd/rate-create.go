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
		monthParam, _ := cmd.Flags().GetString("month")
		if !validateYearMonthFormat(monthParam) {
			return fmt.Errorf("month param  \"%s\" is not valid month format (YYYY-MM)", monthParam)
		}
		fromSymbol, _ := cmd.Flags().GetString("from-symbol")
		toSymbol, _ := cmd.Flags().GetString("to-symbol")
		fromSymbol = strings.ToUpper(fromSymbol)
		toSymbol = strings.ToUpper(toSymbol)
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
		var toSymbolFound bool
		for _, curr := range currencies {
			if fromSymbol == curr.Symbol {
				fromSymbolFound = true
			}
			if toSymbol == curr.Symbol {
				toSymbolFound = true
			}
		}
		if !fromSymbolFound {
			return fmt.Errorf("fromSymbol \"%s\" not found. You must create it first with trackit currency create", fromSymbol)
		}
		if !toSymbolFound {
			return fmt.Errorf("toSymbol \"%s\" not found. You must create it first with trackit currency create", toSymbol)
		}
		if !validateYearMonthFormat(monthParam) {
			return fmt.Errorf("month param \"%s\" must be in form YYYY-MM", monthParam)
		}
		return nil
	},
}

func init() {
	rateCreateCmd.Flags().StringP("month", "m", "", "month in YYYY-MM format for the rate")
	rateCreateCmd.MarkFlagRequired("month")
	rateCreateCmd.Flags().StringP("from-symbol", "f", "", "from currency symbol")
	rateCreateCmd.MarkFlagRequired("from-symbol")
	rateCreateCmd.Flags().StringP("to-symbol", "t", "", "to currency symbol")
	rateCreateCmd.MarkFlagRequired("to-symbol")
	rateCreateCmd.Flags().Float64P("rate", "r", 1, "rate")
	rateCreateCmd.MarkFlagRequired("rate")
	rateCmd.AddCommand(rateCreateCmd)
}
