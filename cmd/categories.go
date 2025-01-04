/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// categoriesCmd represents the categories command
var categoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "Root command for managing categories",
	Long:  `Root command for managing categories, including list, create, update, delete`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("categories called")
	},
}

func init() {
	rootCmd.AddCommand(categoriesCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// categoriesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// categoriesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
