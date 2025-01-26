/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var rateListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the conversion rates",
	Long:  `Lists the the conversion rates. trackit rate list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		month, _ := cmd.Flags().GetString("month")
		db, err := getDB()
		if err != nil {
			return err
		}
		ctx := context.Background()
		queries := models.New(db)
		var rates []models.ReadAllRatesRow
		if month == "" {
			rates, err = queries.ReadAllRates(ctx)
			if err != nil {
				return fmt.Errorf("error reading rates: %w", err)
			}
		} else {
			if !validateYearMonthFormat(month) {
				return errors.New("error parsing month parameter. Should be in format YYYY-MM")
			}
			ratesByMonth, err := queries.ReadRatesByMonth(ctx, month)
			if err != nil {
				return fmt.Errorf("error reading rates by month: %w", err)
			}
			for _, r := range ratesByMonth {
				rates = append(rates, models.ReadAllRatesRow(r))
			}
		}
		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "Month", "From", "To", "Rate"})
		for _, rate := range rates {
			t.AppendRow([]interface{}{rate.ID, rate.Month, rate.FromCurrencySymbol, rate.ToCurrencySymbol, rate.Rate})
		}
		t.Render()
		return nil

	},
}

func init() {
	rateListCmd.Flags().StringP("month", "m", "", "month in YYYY-MM format for the rate")
	rateCmd.AddCommand(rateListCmd)
}
