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

var categoryCreateCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"add"},
	Args:    cobra.ExactArgs(1),
	Short:   "Creates a category. categories add <name>",
	Long:    `Creates a category, taking one positional argument (category name). categories add <name>`,
	RunE: func(cmd *cobra.Command, args []string) error {
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
		err = queries.CreateCategory(ctx, args[0])
		if err != nil {
			return fmt.Errorf("error creating category: %w", err)
		}
		return nil
	},
}

func init() {
	categoryCmd.AddCommand(categoryCreateCmd)
}
