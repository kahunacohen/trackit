/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var currencyUpdateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.ExactArgs(2),
	Short: "Updates an existing currency symbol. trackit currency update <old> <new>",
	Long:  `Updates an existing currency symbol. trackit currency update <old> <new>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fromCurr := args[0]
		toCurr := args[1]
		if len(fromCurr) != 3 || len(toCurr) != 3 {
			return errors.New("from and to currency must have three letter length")
		}
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		err = queries.UpdateCurrencyCode(ctx, models.UpdateCurrencyCodeParams{
			Newsymbol: strings.ToUpper(toCurr), Oldsymbol: strings.ToUpper(fromCurr)})
		if err != nil {
			return fmt.Errorf("error updating currency symbol: %w", err)
		}
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyUpdateCmd)
}
