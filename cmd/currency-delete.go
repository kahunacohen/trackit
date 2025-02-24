/*
Copyright © 2025 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var currencyDeleteCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(1),
	Short: "Deletes a currency symbol by name",
	Long:  `Deletes a currency symbol by name. trackit currency delete ILS`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		err = queries.DeleteCurrencyCode(ctx, args[0])
		if err != nil {
			return fmt.Errorf("error deleting currency code: %w", err)
		}
		return nil
	},
}

func init() {
	currencyCmd.AddCommand(currencyDeleteCmd)
}
