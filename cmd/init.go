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

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"golang.org/x/exp/maps"

	_ "github.com/golang-migrate/migrate/v4/source/file" // Add this!
	_ "github.com/mattes/migrate/source/file"

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

		dbCachePath, err := getDBPathCache()
		if err != nil {
			return err
		}
		userConfigFile, err := os.OpenFile(dbCachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open config file: %w", err)
		}
		defer userConfigFile.Close()
		logF(verbose, "perisisting DB path at: %s with %s", dbCachePath, dbFilePath)
		_, err = userConfigFile.WriteString(dbFilePath)
		if err != nil {
			return fmt.Errorf("failed to write db-path to config file: %w", err)
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
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("error creating DB transaction: %w", err)
		}
		defer db.Close()
		// @TODO pass tx
		// if err = initSchema(db); err != nil {
		// 	return fmt.Errorf("error initializing database schema: %w", err)
		// }

		driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
		if err != nil {
			return fmt.Errorf("error getting driver: %w", err)
		}
		m, err := migrate.NewWithDatabaseInstance("file:///home/aharonc/trackit/internal/db/migrations", "sqlite3", driver)
		if err != nil {
			return fmt.Errorf("error instantiating migration object: %w", err)
		}
		if err := m.Up(); err != nil {
			return fmt.Errorf("error migrating database: %w", err)
		}

		logLn("migrated schema", verbose)
		// @TODO pass tx
		if err = initAccounts(conf, db); err != nil {
			return fmt.Errorf("error initializing accounts: %w", err)
		}

		queries := models.New(tx)
		ctx := context.Background()
		_, err = queries.ReadSettingByName(ctx, "data-dir")
		if err != nil {
			if err == sql.ErrNoRows {
				// Only create setting if it doesn't already exist.
				err = queries.CreateSetting(ctx, models.CreateSettingParams{Name: "data-dir", Value: dataPath})
			} else {
				return fmt.Errorf("error reading data-dir from settings")
			}
		}
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error setting data path: %w", err)
		}
		err = queries.CreateSetting(ctx,
			models.CreateSettingParams{Name: "config-file", Value: configFilePath})
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error writing config-file path to db: %w", err)
		}
		logLn("initialized accounts", verbose)

		if err = initCategories(ctx, conf, queries); err != nil {
			tx.Rollback()
			return fmt.Errorf("error initializing categories: %w", err)
		}
		logLn("initialized categories", verbose)

		if err := queries.CreateSetting(ctx, models.CreateSettingParams{Name: "version", Value: cmd.Version}); err != nil {
			tx.Rollback()
			return fmt.Errorf("error setting version in DB: %w", err)
		}
		logLn("succesfully completed initialization", verbose)
		return nil
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	initCmd.Flags().StringP("data-path", "d", filepath.Join(homeDir, "trackit-data"),
		"Specify the desired path to the directory holding the downloaded CSVs.")
	initCmd.Flags().StringP("db-path", "p", filepath.Join(homeDir, "trackit-data", "trackit.db"),
		"Specify the desired path to the trackit.db (sqlite) database file, including the name of the file")
	initCmd.Flags().StringP("config-file", "c", filepath.Join(homeDir, "trackit-data", "trackit.yaml"),
		"Specify the path to the trackit.yaml config file, including the name of the file")
	rootCmd.AddCommand(initCmd)
}

// Initialize the schema by embedding the schema file (which sqlc also uses)
// and executing it. Because the embedded schema file will only work at the current
// directory, not in the internal/db directory from this go module, the build process must
// copies the schema.sql file to this directory.
// func initSchema(db *sql.DB) error {
// 	if _, err := db.Exec(schemaSQL); err != nil {
// 		return err
// 	}
// 	return nil
// }

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
