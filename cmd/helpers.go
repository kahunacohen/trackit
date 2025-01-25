package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kahunacohen/trackit/internal/models"
)

func roundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func getDB() (*sql.DB, error) {
	path, err := getDBPath()
	if err != nil {
		return nil, err
	}
	if path == nil {
		return nil, errors.New("path to database is nil")
	}
	return sql.Open("sqlite", *path)
}

func getDBPath() (*string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, fmt.Errorf("can't find user cache dir: %w", err)
	}
	cachePath := filepath.Join(cacheDir, "trackit", "cache")
	bytes, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %w", cachePath, err)
	}
	s := string(bytes)
	return &s, nil
}

func renderTransactionTable(rows []models.TransactionsView, total *float64) error {
	t := table.NewWriter()
	t.SetStyle(table.StyleLight)
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"ID", "Date", "Payee", "Account", "Category", "Ignore", "Amount"})
	for _, row := range rows {
		var category string
		if row.CategoryName.Valid {
			category = row.CategoryName.String
		} else {
			category = "-"
		}
		ignoreVal := "No"
		if row.IgnoreWhenSumming == 1 {
			ignoreVal = "Yes"
		}
		t.AppendRow([]interface{}{row.TransactionID, row.Date, row.CounterParty, accountKeyToName(row.AccountName), category, ignoreVal, fmt.Sprintf("%.2f", row.Amount)})
	}
	totalStr := "0.00"
	if total != nil {
		totalStr = strconv.FormatFloat(*total, 'f', 2, 64) // 'f' for floating-point format, 2 digits after the decimal
	}

	t.AppendFooter(table.Row{"", "", "", "", "", "Total", totalStr})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:  "Amount",
			Align: 4,
		},
	})
	t.Render()
	return nil
}

func validateYearMonthFormat(s string) bool {
	_, err := time.Parse("2006-01", s)
	return err == nil
}

func validateDateWithDayFormat(name string) bool {
	split := strings.Split(name, "-")
	if len(split) != 3 {
		return false
	}
	day := split[2]
	d, err := strconv.Atoi(day)
	if err != nil {
		return false
	}
	if d < 1 || d > 31 {
		return false
	}
	month := split[1]
	m, err := strconv.Atoi(month)
	if err != nil {
		return false
	}
	if m < 1 || m > 12 {
		return false
	}
	year := split[0]
	return len(year) == 4
}

func accountKeyToName(account sql.NullString) string {
	if !account.Valid {
		return "-"
	}
	var name string
	split := strings.Split(account.String, "_")
	for i, s := range split {
		name += s
		if i != len(split)-1 {
			name += " "
		}
	}
	return strings.Title(name)
}
