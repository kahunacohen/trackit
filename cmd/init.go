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

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "intialize db",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")

		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		log.Println("created database")
		defer db.Close()
		conf, err := config.ParseConfig("./trackit.yaml")
		if err != nil {
			log.Fatal(err)
		}
		log.Println("parsed configuration file")
		if err = database.InitSchema(conf, db); err != nil {
			log.Fatalf("error initializing schema: %v", err)
		}
		log.Println("initialized schema")
		if err = database.InitAccounts(conf, db); err != nil {
			log.Fatalf("error initializing accounts: %v", err)
		}
		log.Println("initialized accounts")

		if err = database.InitCategories(conf, db); err != nil {
			log.Fatalf("error initializing categories: %v", err)
		}
		log.Println("initialized categories")
		// if err = database.ProcessFiles(conf, db); err != nil {
		// 	log.Fatalf("error initializing accounts: %v", err)
		// }
		// log.Println("initialized transactions")
		log.Println("succesfully completed initialization")
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
