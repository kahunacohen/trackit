/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"golang.org/x/exp/maps"

	_ "github.com/golang-migrate/migrate/v4/source/file" // Add this!
	_ "github.com/mattes/migrate/source/file"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Intializes the internal sqlite database and parses the config.",
	Long: `Initializes (or registers an existing) internal sqlite database and parses the configuration.
You should call trackit init when moving an existing database (trackit.db) to another machine. If
the database file is in a different location than the default (~/trackit-data), then specify the -p flag
to point to where the trackit.db is.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ = rootCmd.PersistentFlags().GetBool("verbose")
		_, configPath, dbPath, err := getDataPaths()
		if err != nil {
			return err
		}
		conf, err := config.ParseConfig(configPath)
		if err != nil {
			return err
		}

		logLn("parsed configuration file", verbose)
		db, err := getDB(dbPath)
		if err != nil {
			return err
		}
		logLn("created database", verbose)
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error creating DB transaction: %w", err)
		}
		defer db.Close()
		if err != nil {
			return fmt.Errorf("error running migrations: %w", err)
		}

		logLn("migrated schema", verbose)
		// @TODO pass tx
		if err = initAccounts(conf, db); err != nil {
			return fmt.Errorf("error initializing accounts: %w", err)
		}

		queries := models.New(tx)
		ctx := context.Background()
		logLn("initialized accounts", verbose)

		if err = initCategories(ctx, conf, queries); err != nil {
			tx.Rollback()
			return fmt.Errorf("error initializing categories: %w", err)
		}
		logLn("initialized categories", verbose)
		logLn("succesfully completed initialization", verbose)
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("error committing transaction: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

// Returns a triplet of trackit data dir, trackit.yaml, and trackit.db
func getDataPaths() (string, string, string, error) {
	dirPath := os.Getenv("TRACKIT_DATA")
	if dirPath == "" {
		return "", "", "", errors.New("must set TRACKIT_DATA environment variable to where your trackit.yaml and account CSV files are")
	}
	dirPath, err := filepath.Abs(dirPath)
	if err != nil {
		return "", "", "", fmt.Errorf("error getting data paths: %w", err)
	}
	configPath := filepath.Join(dirPath, "trackit.yaml")
	dbPath := filepath.Join(dirPath, "trackit.db")
	return dirPath, configPath, dbPath, nil
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

func initCategories(ctx context.Context, conf *config.Config, queries *models.Queries) error {
	for _, category := range maps.Keys(conf.Categories) {
		if err := queries.CreateCategory(ctx, category); err != nil {
			return err
		}

	}
	return nil
}
