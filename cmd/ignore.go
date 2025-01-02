/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var ignoreCmd = &cobra.Command{
	Use:   "ignore",
	Short: "Mark a transaction by ID as ignored when summing",
	Long:  `Mark a specific transaction to ignore when summing or aggregating.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(ids) == 0 {
			return fmt.Errorf("must specify at least one transaction id")
		}
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		for _, id := range ids {
			err := queries.UpdateTransactionIgnore(ctx, models.UpdateTransactionIgnoreParams{ID: id, IgnoreWhenSumming: 1})
			if err != nil {
				return fmt.Errorf("error updating transaction: %w", err)
			}
		}
		return nil
	},
}
var ids []int64

func init() {
	ignoreCmd.Flags().Int64SliceVar(&ids, "id", []int64{}, "Specify multiple IDs")
	rootCmd.AddCommand(ignoreCmd)
}
