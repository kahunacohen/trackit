/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"golang.org/x/exp/maps"

	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Intializes the internal sqlite database and parses the config.",
	Long: `Initializes the internal sqlite database and parses the configuration. Saves
the path to the config file in the database and caches the path to the database file in a
cache file in the user cache directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ := cmd.Flags().GetBool("verbose")
		dataPath, _ := cmd.Flags().GetString("data-path")
		dataPath, err := filepath.Abs(dataPath)
		if err != nil {
			return fmt.Errorf("error getting absolute path of data directory")
		}

		dbFilePath, _ := cmd.Flags().GetString("db-path")
		dbFilePath, err = filepath.Abs(dbFilePath)
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
			return fmt.Errorf("error getting absolute path from passed config-file: %w", err)
		}
		conf, err := config.ParseConfig(configFilePath)
		if err != nil {
			return err
		}

		logLn("parsed configuration file", verbose)
		db, err := getDB()
		if err != nil {
			return err
		}
		logLn("created database", verbose)
		defer db.Close()
		if err = initSchema(db); err != nil {
			return fmt.Errorf("error initializing database schema: %w", err)
		}
		logLn("initialized schema", verbose)
		if err = initAccounts(conf, db); err != nil {
			return fmt.Errorf("error initializing accounts: %w", err)
		}

		queries := models.New(db)
		// @TODO context
		ctx := context.Background()
		err = queries.CreateSetting(ctx, models.CreateSettingParams{Name: "data-dir", Value: dataPath})
		if err != nil {
			return fmt.Errorf("error setting data path: %w", err)
		}
		err = queries.CreateSetting(ctx,
			models.CreateSettingParams{Name: "config-file", Value: configFilePath})
		if err != nil {
			return fmt.Errorf("error writing config-file path to db: %w", err)
		}
		logLn("initialized accounts", verbose)

		if err = initCategories(conf, db); err != nil {
			return fmt.Errorf("error initializing categories: %w", err)
		}
		logLn("initialized categories", verbose)
		logLn("succesfully completed initialization", verbose)
		return nil
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	initCmd.Flags().StringP("data-path", "a", filepath.Join(homeDir, "trackit-data"),
		"Specify the desired path to the directory holding the downloaded CSVs.")
	initCmd.Flags().StringP("db-path", "d", filepath.Join(homeDir, "trackit.db"),
		"Specify the desired path to the trackit.db (sqlite) database file, including the name of the file")
	initCmd.Flags().StringP("config-file", "c", homeDir+"/trackit.yaml",
		"Specify the path to the trackit.yaml config file, including the name of the file")
	rootCmd.AddCommand(initCmd)
}

//go:embed schema.sql
var schemaSQL string

// Initialize the schema by embedding the schema file (which sqlc also uses)
// and executing it. Because the embedded schema file will only work at the current
// directory, not in the internal/db directory from this go module, the build process must
// copies the schema.sql file to this directory.
func initSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	return nil
}

func initAccounts(conf *config.Config, db *sql.DB) error {
	for accountName := range conf.Accounts {
		// Does the account exist already? If not, insert it
		var count int
		query := "SELECT COUNT(*) FROM accounts WHERE name = ?"
		err := db.QueryRow(query, accountName).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := db.Exec("INSERT INTO accounts (name, currency) VALUES (?, ?)",
				accountName, conf.Accounts[accountName].Currency)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func initCategories(conf *config.Config, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, category := range maps.Keys(conf.Categories) {
		_, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
