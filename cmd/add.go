/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/kahunacohen/trackit/internal/config"
	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Adds transactions",
	Long: `Adds transactions by parsing CSV files in the data directory. This will
not parse files whose transactions that already have been added.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := database.GetDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		queries := models.New(db)
		configPath, err := queries.ReadSettingByName(context.Background(), "config-file")
		if err != nil {
			return fmt.Errorf("error getting config-file path from db: %w", err)
		}

		conf, err := config.ParseConfig(configPath)
		if err != nil {
			return err
		}
		err = database.ProcessFiles(conf, db)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
