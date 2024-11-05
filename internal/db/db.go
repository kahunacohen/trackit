package db

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kahunacohen/trackit/internal/config"
)

func GetDB(pathToDBFile string) (*sql.DB, error) {
	return sql.Open("sqlite", pathToDBFile)
}

func InitSchema(conf *config.Config, db *sql.DB) error {
	// Delete existing db if it already exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	pathToDb := path.Join(homeDir, "trackit.db")
	fmt.Println(pathToDb)
	if err, _ := os.Stat(pathToDb); err != nil {
		fmt.Println("exists!")
		if err := os.Remove(path.Join(homeDir, "trackit.db")); err != nil {
			return err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	createAccountTableSQL := `CREATE TABLE IF NOT EXISTS accounts 
	(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL);`
	if _, err := tx.Exec(createAccountTableSQL); err != nil {
		return err
	}

	// createFileTableSQL := `CREATE TABLE IF NOT EXISTS files
	// 	(hash TEXT PRIMARY KEY, name TEXT NOT NULL);`

	// if _, err := tx.Exec(createFileTableSQL); err != nil {
	// 	tx.Rollback()
	// 	return err
	// }

	createTransactionTableSQL := `CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	account_id INTEGER NOT NULL,
	category_id INTEGER NOT NULL,
	counter_party TEXT NOT NULL,
	amount REAL NOT NULL,
	date DATETIME NOT NULL,
	FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
	FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);`

	if _, err := tx.Exec(createTransactionTableSQL); err != nil {
		tx.Rollback()
		return err
	}
	createCategoryTableSQL := `CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT
	);`
	if _, err := tx.Exec(createCategoryTableSQL); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

type Transaction struct {
	Amount       float64
	Category     string
	CounterParty string
	Date         string
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
func validateFileName(fileName string, conf *config.Config) bool {
	for accountName := range conf.Accounts {
		// targetFileName := strings.Replace(strings.ToLower(accountName), " ", "_", -1) + ".csv"
		if accountName+".csv" == fileName {
			return true
		}
	}
	return false
}

func InitAccounts(conf *config.Config, db *sql.DB) error {
	for accountName := range conf.Accounts {
		// Does the account exist already? If not, insert it
		var count int
		query := "SELECT COUNT(*) FROM accounts WHERE name = ?"
		err := db.QueryRow(query, accountName).Scan(&count)
		if err != nil {
			return err
		}
		if count == 0 {
			_, err := db.Exec("INSERT INTO accounts (name) VALUES (?)", accountName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func parseDate(date string) (*string, error) {
	layout := "01/02/2006"
	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, err
	}
	tStr := t.Format("2006-01-02 15:04:05")
	return &tStr, nil
}
func parseAmount(amount string) (*float64, error) {
	ret, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// Returns a map of this form:
// {accountName: {"transaction_date": 0, "counter_party": 3, "amount": 4}}
// This dymamically tells us for each account what column index in the CSV file maps to what
// database table.
func getColIndexes(conf *config.Config) map[string]map[string]int {
	accountToColIndices := make(map[string]map[string]int)
	for accountName, account := range conf.Accounts {
		colIndexMap := make(map[string]int)
		for i, headerMap := range account.Headers {
			tableName := headerMap["table"]
			if tableName == "transaction_date" || tableName == "counter_party" || tableName == "amount" {
				colIndexMap[tableName] = i
			}
		}
		accountToColIndices[accountName] = colIndexMap
	}
	return accountToColIndices
}

func InitTransactions(conf *config.Config, db *sql.DB) error {
	// First create a map of account name to db table names to indices
	// like: {bank_of_america: {date: 0}} etc.
	// so we can know each bank account's csv structure.
	accountsToColIndices := conf.ColIndices()
	dateEntries, err := os.ReadDir(conf.Data)
	if err != nil {
		return err
	}
	for _, dateEntry := range dateEntries {
		dateName := dateEntry.Name()
		validName := validateDateDir(dateName)
		if !validName {
			return fmt.Errorf("month directory '%s' is invalid. Must be mm-yyyy", dateName)
		}
		monthPath := filepath.Join(conf.Data, dateName)
		fileEntries, err := os.ReadDir(monthPath)
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
			filePath := filepath.Join(monthPath, fileName)
			file, err := os.Open(filePath)
			if err != nil {
				return fmt.Errorf("error opening %s: %w", filePath, err)
			}
			defer file.Close()
			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				return fmt.Errorf("error reading %s: %w", filePath, err)
			}
			if len(records) < 2 {
				return fmt.Errorf("there are less than 2 rows for file: %s", fileName)
			}
			headersInFile := records[0]
			dataRows := records[1:]
			if len(dataRows) == 0 {
				return fmt.Errorf("file %s has no records", filePath)
			}
			bankAccountNameFromFile := strings.Replace(fileName, ".csv", "", 1)
			headersInConfig := conf.Headers(bankAccountNameFromFile)
			colIndices := accountsToColIndices[bankAccountNameFromFile]

			// check if headers config at least all exist in file headers. It could
			// be that there are headers in the file that don't exist in config and
			// and that's ok.
			for _, headerInConfig := range headersInConfig {
				if !slices.Contains(headersInFile, headerInConfig) {
					return fmt.Errorf("header '%s' in file: '%s' is not a valid header for this account: Check trackit.yaml", headerInConfig, filePath)
				}
			}

			// Category is always the last column.
			categoryIndex := len(dataRows[0]) - 1
			for _, row := range dataRows {
				date, err := parseDate(row[colIndices["transaction_date"]])
				if err != nil {
					return fmt.Errorf("error parsing date %s: %v", *date, err)
				}
				amount, err := parseAmount(row[colIndices["amount"]])
				if err != nil {
					return fmt.Errorf("error parsing amount: %f", *amount)
				}
				transaction := Transaction{Date: *date, Amount: *amount, CounterParty: row[colIndices["counter_party"]], Category: row[categoryIndex]}
				fmt.Println(transaction)
			}

		}

	}
	return nil
}
