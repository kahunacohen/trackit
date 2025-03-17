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

var categoryUpdateCmd = &cobra.Command{
	Use:   "update",
	Args:  cobra.ExactArgs(2),
	Short: "Updates an existing category. trackit category update <old> <new>",
	Long:  `Updates an existing category. trackit category update <old> <new>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fromCat := args[0]
		toCat := args[1]
		_, _, dbPath, err := getDataPaths()
		if err != nil {
			return err
		}
		db, err := getDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		queries := models.New(db)
		err = queries.UpdateCategory(ctx, models.UpdateCategoryParams{
			Newcategory: toCat, Oldcategory: fromCat})
		if err != nil {
			return fmt.Errorf("error updating category: %w", err)
		}
		return nil
	},
}

func init() {
	categoryCmd.AddCommand(categoryUpdateCmd)
}
