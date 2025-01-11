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

// categoriesAdd.goCmd represents the categoriesAdd.go command
var categoriesCreateCmd = &cobra.Command{
	Use:   "create",
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is passed
	Short: "Creates a category. categories add <name>",
	Long:  `Creates a category`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
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
	categoriesCmd.AddCommand(categoriesCreateCmd)
}
