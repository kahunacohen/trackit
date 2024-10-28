package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	m, err := strconv.Atoi(month)
	if err != nil {
		return false
	}
	if m < 1 || m > 12 {
		return false
	}
	year := split[1]
	return len(year) == 4
}

func validateFileName(name string, conf *config.Config) bool {
	targetFileName := strings.Replace(strings.ToLower(conf.Accounts[0].Name), " ", "_", -1) + ".csv"
	fmt.Println(targetFileName)
	var found bool
	for _, account := range conf.Accounts {

	}
	return true
}

func InitTransactions(conf *config.Config, db *sql.DB) error {
	dateEntries, err := os.ReadDir(conf.Data)
	if err != nil {
		return err
	}
	// transactionData := TransactionData{}
	for _, dateEntry := range dateEntries {
		dateName := dateEntry.Name()
		validName := validateDateDir(dateName)
		if !validName {
			return fmt.Errorf("Month directory '%s' is invalid. Must be mm-yyyy", dateName)
		}
		fileEntries, err := os.ReadDir(filepath.Join(conf.Data, dateName))
		if err != nil {
			return err
		}
		if len(fileEntries) == 0 {
			return fmt.Errorf("month directory '%s' contains no CSV files", dateName)
		}
		for _, fileEntry := range fileEntries {
			fileName := fileEntry.Name()
			validFileName := validateFileName(fileName, conf)
			if !validFileName {
				return fmt.Errorf("file name '%s' is invalid: it must be a name of a bank account (with spaces separated by '_') defined in trackit.yaml with a .csv extension", fileName)
			}
		}

	}
	return nil
}
