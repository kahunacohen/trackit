/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	Run: func(cmd *cobra.Command, args []string) {
		s, _ := database.GetCachedDbPath()
		fmt.Println(*s)

		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")

		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		queries := models.New(db)
		configPath, err := queries.ReadSettingByName(context.Background(), "config-file")
		if err != nil {
			log.Fatalf("error getting config-file path from db: %v", err)
		}

		conf, err := config.ParseConfig(configPath)
		if err != nil {
			log.Fatal(err)
		}
		err = database.ProcessFiles(conf, db)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
