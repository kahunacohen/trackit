/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

// currencyListCmd represents the currencyList command
var currencyListCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists the currencies registered in the system",
	Long:  `Lists the ISO 4217 currency symbols registered in the system.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		codes, err := queries.ReadCurrencyCodes(ctx)
		if err != nil {
			return fmt.Errorf("error reading currency codes: %w", err)
		}
		fmt.Println(codes)
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyListCmd)
}
