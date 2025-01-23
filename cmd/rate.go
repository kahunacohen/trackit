/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// rateCmd represents the rate command
var rateCmd = &cobra.Command{
	Use:   "rate",
	Short: "root command for rate",
	Long:  `root command for rate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("must call a sub command")
	},
}

func init() {
	rootCmd.AddCommand(rateCmd)
}
