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

	database "github.com/kahunacohen/trackit/internal/db"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

// categorizeCmd represents the categorize command
var categorizeCmd = &cobra.Command{
	Use:   "categorize",
	Short: "Categorizes transactions",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, _ := os.UserHomeDir()
		dbPath := filepath.Join(homeDir, "trackit.db")
		db, err := database.GetDB(dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		ctx := context.Background()
		if interactive {
			queries := models.New(db)
			rows, err := queries.ReadNonCategorizedTransactions(ctx)
			if err != nil {
				return fmt.Errorf("error reading non categorized transactions: %w", err)
			}
			items := []string{"foo", "bar"}

			for _, row := range rows {
				prompt := promptui.Select{
					Label: fmt.Sprintf("Please select a category for %s | %s | %f on %s", row.AccountName, row.CounterParty, row.Amount, row.Date),
					Items: items,
				}
				_, result, err := prompt.Run()
				if err != nil {
					log.Fatalf("Prompt failed %v", err)
				}

				// Print the selected option
				fmt.Printf("You chose: %s\n", result)
			}

		}
		return nil
	},
}
var interactive bool

func init() {
	categorizeCmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "Enable interactive mode")
	rootCmd.AddCommand(categorizeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// categorizeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// categorizeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
