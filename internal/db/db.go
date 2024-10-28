package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/kahunacohen/trackit/internal/config"
)

func GetDB(pathToDBFile string) (*sql.DB, error) {
	return sql.Open("sqlite", pathToDBFile)
}

func InitSchema(accounts []config.Account, db *sql.DB) error {
	createAccountTableSQL := `CREATE TABLE IF NOT EXISTS accounts 
		(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);`
	if _, err := db.Exec(createAccountTableSQL); err != nil {
		return err
	}
	createTransactionTableSQL := `CREATE TABLE IF NOT EXISTS transactions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		account_id INTEGER NOT NULL,
		category_id INTEGER NOT NULL,
		description TEXT,
		amount REAL NOT NULL,
		transaction_date DATETIME NOT NULL,
		FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
		FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
	);`
	if _, err := db.Exec(createTransactionTableSQL); err != nil {
		return err
	}
	createCategoryTableSQL := `CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	);`
	if _, err := db.Exec(createCategoryTableSQL); err != nil {
		return err
	}
	return nil
}

type Transaction struct {
	Date        string
	Description *string
	Amount      float64
	Category    *string
}
type Account struct {
	Name         string
	Transactions []Transaction
}
type Month struct {
	YearMonth string
	Accounts  []Account
}
type TransactionData struct {
	Months []Month
}

func validateDateDir(name string) bool {
	split := strings.Split(name, "-")
	if len(split) != 2 {
		return false
	}
	month := split[0]
	year := split[1]
	// validate month and year.
	return true

}
func InitTransactions(conf *config.Config, db *sql.DB) error {
	entries, err := os.ReadDir(conf.Data)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		validName := validateDateDir(name)
		if !validName {
			return fmt.Errorf("Month directory '%s' is invalid. Must be mm-yyyy", name)
		}
	}
	return nil
}
