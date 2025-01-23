/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var currencyCmd = &cobra.Command{
	Use:   "currency",
	Short: "Root command for currency management",
	Long:  `Root command for currency management including create, read, update, delete for currency symbols.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must call a subcommand")
	},
}

func init() {
	rootCmd.AddCommand(currencyCmd)
}
