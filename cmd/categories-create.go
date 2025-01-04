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

// categoriesAdd.goCmd represents the categoriesAdd.go command
var categoriesCreateCmd = &cobra.Command{
	Use:   "create",
	Args:  cobra.ExactArgs(1), // Ensure exactly one argument is passed
	Short: "Creates a category. categories add <name>",
	Long:  `Creates a category`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// categoriesAdd.goCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// categoriesAdd.goCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
