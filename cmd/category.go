/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// categoryCmd represents the categories command
var categoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Root command for managing categories",
	Long:  `Root command for managing categories, including list, create, update, delete`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must call a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(categoryCmd)
}
