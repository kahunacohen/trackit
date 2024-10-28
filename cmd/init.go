/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"log"
	"os"
	"path/filepath"

	"github.com/kahunacohen/trackit/internal/config"

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/spf13/cobra"
	_ "modernc.org/sqlite"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize tracking.",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")

		// Open a SQLite database file
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		conf, err := config.ParseConfig("./trackit.yaml")
		if err != nil {
			log.Fatal(err)
		}
		if err = database.InitSchema(conf.Accounts, db); err != nil {
			log.Fatalf("error initializing accounts: %v", err)
		}
		if err = database.InitTransactions(conf, db); err != nil {
			log.Fatalf("error initializing accounts: %v", err)
		}

		// Use the database (e.g., creating tables, inserting data, querying)
		// if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS expenses (id INTEGER PRIMARY KEY, amount REAL, category TEXT);`); err != nil {
		// 	log.Fatalf("Failed to create table: %v", err)
		// }
	},
}

func init() {
	rootCmd.AddCommand(initCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// initCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// initCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
