/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"errors"
	"strconv"

	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
)

var rateDeleteCmd = &cobra.Command{
	Use:   "delete",
	Args:  cobra.ExactArgs(1),
	Short: "Deletes a rate by ID.",
	Long: `Deletes a rate by ID. Get a list of rates with trackit rate list.
Delete a rate: trackit delete <id>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rateID, err := strconv.Atoi(args[0])
		if err != nil {
			return errors.New("could not parse argument as a rate ID")
		}
		db, _, err := getDB()
		if err != nil {
			return err
		}
		ctx := context.Background()
		queries := models.New(db)
		queries.DeleteRate(ctx, int64(rateID))
		return nil
	},
}

func init() {
	rateCmd.AddCommand(rateDeleteCmd)
}
