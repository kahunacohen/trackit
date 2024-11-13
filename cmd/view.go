/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"regexp"

	"github.com/spf13/cobra"
)

func validateDateFormat(date string) (bool, error) {
	// Regular expression for MM-YYYY format
	re := regexp.MustCompile(`^(0[1-9]|1[0-2])-[0-9]{4}$`)
	return re.MatchString(date), nil
}

// viewCmd represents the view command
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Generates a view of account data",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")
		if date == "" && account == "" {
			return fmt.Errorf("either date or account flag is required")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().StringP("date", "d", "", "Date in MM-YYYY format")
	viewCmd.Flags().StringP("account", "a", "", "One of the account names in your trackit config file")
	viewCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Get the date flag value
		date, _ := cmd.Flags().GetString("date")
		account, _ := cmd.Flags().GetString("account")

		// Check if the date is in the correct MM-YYYY format using regex
		if date != "" {
			isValid, err := validateDateFormat(date)
			if err != nil {
				return err
			}
			if !isValid {
				return fmt.Errorf("date must be in MM-YYYY format")
			}
		}
		if account != "" {
			// @TODO get DB, refactor so that db module handles getting db.
			fmt.Println("validate bank account is valid")
		}

		return nil
	}
}
