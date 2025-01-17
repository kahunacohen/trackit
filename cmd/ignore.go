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

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Marks a transaction by ID as ignored when summing",
	Long: `Mark a specific transaction to ignore when summing or aggregating. To undo ignore,
pass the --toggle flag along with --id`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(ids) == 0 {
			return fmt.Errorf("must specify at least one transaction id")
		}
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		for _, id := range ids {
			var valToSet int64 = 1
			if toggle {
				transaction, err := queries.ReadTransactionById(ctx, id)
				if err != nil {
					return fmt.Errorf("error getting transaction to toggle: %w", err)
				}
				if transaction.IgnoreWhenSumming == 1 {
					valToSet = 0
				}
			}

			err := queries.UpdateTransactionIgnore(ctx, models.UpdateTransactionIgnoreParams{ID: id, IgnoreWhenSumming: valToSet})
			if err != nil {
				return fmt.Errorf("error updating transaction: %w", err)
			}
		}
		return nil
	},
}
var ids []int64
var toggle bool

func init() {
	ignoreCmd.Flags().Int64SliceVar(&ids, "id", []int64{}, "Specify multiple IDs")
	ignoreCmd.Flags().BoolVar(&toggle, "toggle", false, "Toggle whether a transaction should be ignored or not when summing. Must be supplied with the --id flag")
	rootCmd.AddCommand(ignoreCmd)
}
