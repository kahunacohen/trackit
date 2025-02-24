/*
Copyright © 2025 Aaron Cohen <aaroncohendev@gmail.com>
*/
package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var transactionDeleteCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"rm"},
	Args:    cobra.ExactArgs(1),
	Short:   "deletes a transaction",
	Long:    `deletes a transaction by ID. trackit transaction delete <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			return err
		}
		queries := models.New(db)
		transactionID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("error converting id to int: %w", err)
		}
		if err = queries.DeleteTransaction(context.Background(), int64(transactionID)); err != nil {
			return fmt.Errorf("error deleting transaction: %w", err)
		}
		return nil
	},
}

func init() {
	transactionCmd.AddCommand(transactionDeleteCmd)
}
