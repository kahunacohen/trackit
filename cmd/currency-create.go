/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var currencyCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a new ISO 4217 currency symbol. trackit currency create ILS",
	Long:    `Creates a new ISO 4217 three-letter ISO symbol. "USD" is pre-populated. trackit currency create ILS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args[0]) != 3 {
			return errors.New("currency symbol must be three characters")
		}
		symbol := strings.ToUpper(args[0])
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		err = queries.CreateCurrencyCode(ctx, symbol)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyCreateCmd)
}
