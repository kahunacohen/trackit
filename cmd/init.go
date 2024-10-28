/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

//go:embed assets/trackit.db
var dbFile []byte

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
		fmt.Println("init called")
		if _, err := os.Stat("./trackit.db"); os.IsNotExist(err) {
			err := os.WriteFile("trackit.db", dbFile, 0644)
			if err != nil {
				log.Fatalf("problem writing to db file: %w", err)
			}
		}
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
