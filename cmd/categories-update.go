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
	"strconv"

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{Use: "update", Args: cobra.ExactArgs(2), Short: "updates an existing category. update <id> <name>", Long: `updates an existing category by id. update <id> <name>`, RunE: func(cmd *cobra.Command, args []string) error {
	homeDir, _ := os.UserHomeDir()
	dbPath := filepath.Join(homeDir, "trackit.db")
	db, err := database.GetDB(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	ctx := context.Background()
	queries := models.New(db)
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("could not update category with id \"%v\"", args[0])
	}
	err = queries.UpdateCategory(ctx, models.UpdateCategoryParams{ID: id, Name: args[1]})
	if err != nil {
		return fmt.Errorf("error updating category: %w", err)
	}
	return nil
}}

func init() {
	categoriesCmd.AddCommand(updateCmd)
}
