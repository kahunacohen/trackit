/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(1),
	Short: "Deletes a category. delete <id>",
	Long: `Deletes an existing category by ID. trackit categories delete <id>. Get the category
id by doing trackit categories list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing id: %w", err)
		}
		err = queries.DeleteCategory(ctx, id)
		if err != nil {
			return fmt.Errorf("error deleting category: %w", err)
		}
		return nil
	},
}

func init() {
	categoriesCmd.AddCommand(deleteCmd)
}
