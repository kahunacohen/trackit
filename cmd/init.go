/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/manifoldco/promptui"
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
		err := generateTrackitYML()
		if err != nil {
			return err
		}

		// verbose, _ = rootCmd.PersistentFlags().GetBool("verbose")
		// dataPath, _ := cmd.Flags().GetString("data-path")
		// dataPath, err := filepath.Abs(dataPath)
		// if err != nil {
		// 	return fmt.Errorf("error getting absolute path of data directory")
		// }

		// dbFilePath, _ := cmd.Flags().GetString("db-path")
		// dbFilePath, err = filepath.Abs(dbFilePath)
		// if err != nil {
		// 	return fmt.Errorf("error getting absolute path for supplied db-path: %w", err)
		// }

		// dbCachePath, err := getDBPathCache()
		// if err != nil {
		// 	return err
		// }
		// userConfigFile, err := os.OpenFile(dbCachePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		// if err != nil {
		// 	return fmt.Errorf("failed to open config file: %w", err)
		// }
		// defer userConfigFile.Close()
		// logF(verbose, "perisisting DB path at: %s with %s", dbCachePath, dbFilePath)
		// _, err = userConfigFile.WriteString(dbFilePath)
		// if err != nil {
		// 	return fmt.Errorf("failed to write db-path to config file: %w", err)
		// }

		// configFilePath, _ := cmd.Flags().GetString("config-file")
		// configFilePath, err = filepath.Abs(configFilePath)
		// if err != nil {
		// 	return fmt.Errorf("error getting absolute path from passed config-file: %w", err)
		// }
		// conf, err := config.ParseConfig(configFilePath)
		// if err != nil {
		// 	return err
		// }

		// logLn("parsed configuration file", verbose)
		// db, err := getDB()
		// if err != nil {
		// 	return err
		// }
		// logLn("created database", verbose)
		// tx, err := db.Begin()
		// if err != nil {
		// 	return fmt.Errorf("error creating DB transaction: %w", err)
		// }
		// defer db.Close()
		// if err != nil {
		// 	return fmt.Errorf("error running migrations: %w", err)
		// }

		// logLn("migrated schema", verbose)
		// // @TODO pass tx
		// if err = initAccounts(conf, db); err != nil {
		// 	return fmt.Errorf("error initializing accounts: %w", err)
		// }

		// queries := models.New(tx)
		// ctx := context.Background()
		// _, err = queries.ReadSettingByName(ctx, "data-dir")
		// if err != nil {
		// 	if err == sql.ErrNoRows {
		// 		// Only create setting if it doesn't already exist.
		// 		err = queries.CreateSetting(ctx, models.CreateSettingParams{Name: "data-dir", Value: dataPath})
		// 	} else {
		// 		return fmt.Errorf("error reading data-dir from settings")
		// 	}
		// }
		// if err != nil {
		// 	tx.Rollback()
		// 	return fmt.Errorf("error setting data path: %w", err)
		// }
		// err = queries.CreateSetting(ctx,
		// 	models.CreateSettingParams{Name: "config-file", Value: configFilePath})
		// if err != nil {
		// 	tx.Rollback()
		// 	return fmt.Errorf("error writing config-file path to db: %w", err)
		// }
		// logLn("initialized accounts", verbose)

		// if err = initCategories(ctx, conf, queries); err != nil {
		// 	tx.Rollback()
		// 	return fmt.Errorf("error initializing categories: %w", err)
		// }
		// logLn("initialized categories", verbose)

		// if err := queries.CreateSetting(ctx, models.CreateSettingParams{Name: "version", Value: cmd.Version}); err != nil {
		// 	tx.Rollback()
		// 	return fmt.Errorf("error setting version in DB: %w", err)
		// }
		// logLn("succesfully completed initialization", verbose)
		// if err := tx.Commit(); err != nil {
		// 	return fmt.Errorf("error committing transaction: %w", err)
		// }
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

func normalizePath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[1:])
	}
	return filepath.Abs(path)
}

func generateTrackitYML() error {
	conf := config.Config{}
	prompt := promptui.Prompt{
		Label: "Enter an existing directory to save the config file (trackit.yaml) to",
	}
	ymlPath, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	ymlPath, err = normalizePath(ymlPath)
	if err != nil {
		return fmt.Errorf("error making yaml path absolute: %w", err)
	}
	stat, err := os.Stat(ymlPath)
	if err == nil && stat.IsDir() {

	} else {
		return fmt.Errorf("%s is doesn't exist or is not a directory", ymlPath)
	}

	prompt = promptui.Prompt{
		Label: "Enter a base currency (USD, ILS etc.). This is the currency all transactions should be converted to.",
	}
	baseCurrency, err := prompt.Run()
	if err != nil {
		return err
	}
	if len(baseCurrency) != 3 {
		return fmt.Errorf("base currency must be 3 characters")
	}
	baseCurrency = strings.ToUpper(baseCurrency)
	conf.BaseCurrency = baseCurrency
	accountsMap := make(map[string]config.Account)

	for {
		prompt = promptui.Prompt{
			Label: "Enter the name of an account  (e.g. Bank of America).",
		}
		accountName, err := prompt.Run()
		if err != nil {
			return err
		}
		if accountName == "" {
			break
		}
		prompt = promptui.Prompt{Label: "Enter the date format this account uses (e.g., \"mm-dd-yyyy\", \"yyyy/mm/dd\""}
		dateLayout, err := prompt.Run()
		if err != nil {
			return err
		}

		prompt = promptui.Prompt{Label: "Enter the currency for this account (e.g. USD, ILS etc.)"}
		bankAccountCurrency, err := prompt.Run()
		if err != nil {
			return err
		}

		prompt = promptui.Prompt{Label: "Enter whether debits are shown as positive (y/n). Default is \"n\"."}
		var debitAsPositiveBool bool
		debitAsPositive, err := prompt.Run()
		if err != nil {
			return err
		}
		debitAsPositive = strings.ToLower(debitAsPositive)
		if debitAsPositive == "y" || debitAsPositive == "yes" {
			debitAsPositiveBool = true
		}

		prompt = promptui.Prompt{Label: "Enter thousands separator for this account (e.g. what character is used to separate thousands place). Enter ~ for none."}
		sep, err := prompt.Run()
		if err != nil {
			return err
		}
		accountsMap[accountNameToKey(accountName)] = config.Account{
			Currency:           bankAccountCurrency,
			DateLayout:         dateLayout,
			DebitAsPositive:    debitAsPositiveBool,
			ThousandsSeparator: sep}
		conf.Accounts = accountsMap
	}
	fmt.Println(conf.WriteToYaml())

	return nil
}

func dateTokenToGoFormat(dateToken string) string {
	format := strings.ToLower(dateToken)
	format = strings.ReplaceAll(format, "yyyy", "2006")
	format = strings.ReplaceAll(format, "yy", "06")
	format = strings.ReplaceAll(format, "mm", "01")
	format = strings.ReplaceAll(format, "dd", "02")
	return format
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
