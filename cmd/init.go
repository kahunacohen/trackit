/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
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
	Short: "Intializes internal database and parses configuration",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		dbFilePath, _ := cmd.Flags().GetString("db-path")
		dbFilePath, err := filepath.Abs(dbFilePath)
		if err != nil {
			log.Fatalf("error getting absolute path for db-path: %v", err)
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
		db, err := database.GetDB(dbFilePath)
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
		// Save config file path to db
		queries := models.New(db)
		// @TODO backround?
		ctx := context.Background()

		// save the path to the db in a local file for later access, not the db because storing
		// the path in the db would create a circular dependency. We need the db to get the setting
		// and need the setting to get the db.
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			log.Fatalf("can't find user cache dir: %v", err)
		}
		cachePath := filepath.Join(cacheDir, "trackit")
		if err := os.MkdirAll(cachePath, 0755); err != nil {
			log.Fatalf("error creating trackit cache directory: %v", err)
		}
		file, err := os.OpenFile(filepath.Join(cachePath, "cache"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatalf("failed to open cache file: %v", err)
		}
		defer file.Close() // Ensure the file is closed when done

		// Write the string to the file
		_, err = file.WriteString(dbFilePath)
		if err != nil {
			log.Fatalf("failed to write to cache file: %v", err)
		}

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
	},
}

func init() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("error getting home directory: %v", err)
	}
	initCmd.Flags().StringP("db-path", "d", homeDir+"/trackit.db",
		"Specify the desired path to the trackit.db (sqlite) database file, including the name of the file")
	initCmd.Flags().StringP("config-file", "c", homeDir+"/trackit.yaml",
		"Specify the path to the trackit.yaml config file, including the name of the file")
	rootCmd.AddCommand(initCmd)
}
