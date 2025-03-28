/*
Copyright © 2025 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var categoryListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "lists existing categories",
	Long:    `lists existing categories. trackit category list`,
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
		categories, err := queries.ReadAllCategories(ctx)
		if err != nil {
			return fmt.Errorf("error getting categories: %w", err)
		}
		t := table.NewWriter()
		t.SetStyle(table.StyleLight)
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"ID", "Name"})
		for _, category := range categories {
			t.AppendRow([]interface{}{category.ID, category.Name})
		}
		t.Render()
		return nil
	},
}

func init() {
	categoryCmd.AddCommand(categoryListCmd)
}
