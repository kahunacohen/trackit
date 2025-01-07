/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Intializes the internal sqlite database and parses the configuration",
	Long: `Initializes the internal sqlite database and parses the configuration. Saves
the path to the config file in the database and caches the path to the database file in a
cache file in the user cache directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dbFilePath, _ := cmd.Flags().GetString("db-path")
		dbFilePath, err := filepath.Abs(dbFilePath)
		if err != nil {
			return fmt.Errorf("error getting absolute path for supplied db-path: %w", err)
		}

		// Save the path to the db in a local cache file for later access. Don't
		// save this to the DB  because storing
		// that would create a circular dependency. We need the db to get the setting
		// and need the setting to get the db.
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			return fmt.Errorf("can't get user cache dir: %w", err)
		}
		cachePath := filepath.Join(cacheDir, "trackit")
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			return fmt.Errorf("error creating trackit cache directory: %w", err)
		}
		file, err := os.OpenFile(filepath.Join(cachePath, "cache"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open cache file: %w", err)
		}
		defer file.Close()
		_, err = file.WriteString(dbFilePath)
		if err != nil {
			return fmt.Errorf("failed to write db-path to cache file: %w", err)
		}

		configFilePath, _ := cmd.Flags().GetString("config-file")
		configFilePath, err = filepath.Abs(configFilePath)
		if err != nil {
			log.Fatalf("error getting absolute path from passed config-file: %v", err)
		}
		conf, err := config.ParseConfig(configFilePath)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("parsed configuration file")
		db, err := database.GetDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		log.Println("created database")
		defer db.Close()
		if err = database.InitSchema(conf, db); err != nil {
			log.Fatalf("error initializing schema: %v", err)
		}
		log.Println("initialized schema")
		if err = database.InitAccounts(conf, db); err != nil {
			log.Fatalf("error initializing accounts: %v", err)
		}

		queries := models.New(db)
		// Create transaction view programatically

		ctx := context.Background()
		// Save config file path to db

		err = queries.CreateSetting(ctx,
			models.CreateSettingParams{Name: "config-file", Value: configFilePath})
		if err != nil {
			log.Fatalf("error writing config-file path to db: %v", err)
		}
		log.Println("initialized accounts")

		if err = database.InitCategories(conf, db); err != nil {
			log.Fatalf("error initializing categories: %v", err)
		}
		log.Println("initialized categories")
		log.Println("succesfully completed initialization")
		return nil
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	initCmd.Flags().StringP("db-path", "d", filepath.Join(homeDir, "trackit.db"),
		"Specify the desired path to the trackit.db (sqlite) database file, including the name of the file")
	initCmd.Flags().StringP("config-file", "c", homeDir+"/trackit.yaml",
		"Specify the path to the trackit.yaml config file, including the name of the file")
	rootCmd.AddCommand(initCmd)
}
