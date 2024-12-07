package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kahunacohen/trackit/internal/config"
	"golang.org/x/exp/maps"
	"sigs.k8s.io/yaml"
)

type ExchangeRate struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

type ExchangeRatesWrapper struct {
	ExchangeRates []ExchangeRate `json:"exchange_rates"`
}

type CategoryAgregation struct {
	Category string
	Total    float64
}

func RoundAmount(amount float64) float64 {
	return math.Round(amount*100) / 100
}

func GetDB(pathToDBFile string) (*sql.DB, error) {
	return sql.Open("sqlite", pathToDBFile)
}

func InitSchema(conf *config.Config, db *sql.DB) error {
	// Delete existing db if it already exists
	// homeDir, err := os.UserHomeDir()
	// if err != nil {
	// 	return err
	// }
	// pathToDb := path.Join(homeDir, "trackit.db")
	// if err, _ := os.Stat(pathToDb); err != nil {
	// 	if err := os.Remove(path.Join(homeDir, "trackit.db")); err != nil {
	// 		return err
	// 	}
	// }

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	createFileTableSQL := `CREATE TABLE IF NOT EXISTS files 
	(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, hash TEXT NOT NULL);`
	if _, err := tx.Exec(createFileTableSQL); err != nil {
		return err
	}

	createAccountTableSQL := `CREATE TABLE IF NOT EXISTS accounts 
	(id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, currency TEXT NOT NULL);`
	if _, err := tx.Exec(createAccountTableSQL); err != nil {
		return err
	}

	createTransactionTableSQL := `CREATE TABLE IF NOT EXISTS transactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	account_id INTEGER NOT NULL,
	category_id INTEGER NOT NULL,
	counter_party TEXT NOT NULL,
	amount REAL NOT NULL,
	deposit REAL,
	withdrawl REAL,
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

	createTransactionViewSQL := `
CREATE VIEW transactions_view AS
SELECT 
    accounts.id AS account_id,
    accounts.name AS account_name, 
    transactions.id AS transaction_id, 
	transactions.date AS date, 
    transactions.counter_party AS counter_party, 
    transactions.amount AS amount, 
    categories.name AS category_name
FROM 
    transactions
JOIN 
    accounts ON transactions.account_id = accounts.id
LEFT JOIN 
    categories ON transactions.category_id = categories.id
ORDER BY date DESC;`
	if _, err := tx.Exec(createTransactionViewSQL); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

type Transaction struct {
	Amount       float64
	Category     *string
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
func GetAccountTransactions(db *sql.DB, accountName string, date string) ([]Transaction, error) {
	var transactions []Transaction
	var rows *sql.Rows
	var err error

	// account and date are not set
	if accountName == "" && date == "" {
		rows, err = db.Query("SELECT date, counter_party, amount, category_name FROM transactions_view;")
	} else if accountName != "" && date != "" { // Both are set
		rows, err = db.Query("SELECT date, counter_party, amount, category_name FROM transactions_view WHERE account_name=? AND strftime('%m-%Y', date)=?",
			accountName, date)
		// account name is set but not date
	} else if accountName != "" && date == "" {
		rows, err = db.Query("SELECT date, counter_party, amount, category_name FROM transactions_view WHERE account_name=?",
			accountName)
		// date is set but not account
	} else {
		rows, err = db.Query("SELECT date, counter_party, amount, category_name FROM transactions_view WHERE strftime('%m-%Y', date)=?",
			date)
	}
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var transaction Transaction
		if err := rows.Scan(&transaction.Date, &transaction.CounterParty, &transaction.Amount, &transaction.Category); err != nil {
			return nil, err // Handle scan error
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
func GetCategoryAggregation(db *sql.DB, account string, date string) ([]CategoryAgregation, error) {
	var rows *sql.Rows
	var err error
	if account == "" && date == "" {
		rows, err = db.Query("SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view GROUP BY category_name ORDER BY total_amount;")
	} else if account != "" && date == "" {
		rows, err = db.Query("SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE account_name=? GROUP BY category_name ORDER BY total_amount;", account)
	} else if account == "" && date != "" {
		rows, err = db.Query("SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE strftime('%m-%Y', date)=? GROUP BY category_name ORDER BY total_amount;", date)
	} else {
		rows, err = db.Query("SELECT COALESCE(category_name, 'uncategorized') AS category_name, SUM(amount) AS total_amount FROM transactions_view WHERE account_name=? AND strftime('%m-%Y', date)=? GROUP BY category_name ORDER BY total_amount;",
			account, date)
	}
	if err != nil {
		return nil, err
	}
	var aggregates []CategoryAgregation
	for rows.Next() {
		var aggregate CategoryAgregation
		if err := rows.Scan(&aggregate.Category, &aggregate.Total); err != nil {
			return nil, err
		}
		aggregate.Total = RoundAmount(aggregate.Total)
		aggregates = append(aggregates, aggregate)
	}
	return aggregates, nil
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
			_, err := db.Exec("INSERT INTO accounts (name, currency) VALUES (?, ?)",
				accountName, conf.Accounts[accountName].Currency)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func InitCategories(conf *config.Config, db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	for _, category := range maps.Keys(conf.Categories) {
		_, err := tx.Exec("INSERT INTO categories (name) VALUES (?)", category)
		if err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func parseDate(layout string, date string) (*string, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return nil, err
	}
	tStr := t.Format("2006-01-02 15:04:05")
	return &tStr, nil
}
func parseAmount(amount string, thousandsSeparator string) (*float64, error) {
	var amountStr string
	if thousandsSeparator != "" {
		amountStr = strings.Replace(amount, thousandsSeparator, "", 1)
	} else {
		amountStr = amount
	}
	ret, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

func computeFileHash(file *os.File) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", fmt.Errorf("error resetting file pointer when trying to compute hash: %w", err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func AddData(conf *config.Config, db *sql.DB) error {
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
			if fileName == "rate.yaml" || fileName == "rate.yml" {
				continue
			}
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

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("error beginning db transaction when inserting transactions: %w", err)
			}
			fileHash, err := computeFileHash(file)
			if err != nil {
				return fmt.Errorf("problem hashing file: %w", err)
			}

			// Check if file has been modified
			var hashFromDb *string
			var fileShouldBeProcessed bool
			log.Printf("checking if file %s has been processed\n", filePath)
			err = db.QueryRow("SELECT hash FROM files where name=?", filePath).Scan(&hashFromDb)
			if err != nil {
				if err == sql.ErrNoRows {
					// The file never had a hash and has never been seen before.
					log.Printf("file %s has never been processed, insert hash to db\n", filePath)
					_, err := db.Exec("INSERT INTO files (name, hash) VALUES (?, ?)", filePath, fileHash)
					hashFromDb = &fileHash
					if err != nil {
						return fmt.Errorf("error inserting initial file hash: %w", err)
					}
					fileShouldBeProcessed = true
				} else {
					// Some other error trying to get hash.
					return fmt.Errorf("error looking up hash from db for %s: %v", filePath, err)
				}
			}
			if hashFromDb == nil {
				return fmt.Errorf("hash from db for file %s is nil", filePath)
			}
			if fileHash != *hashFromDb {
				fileShouldBeProcessed = true
			}
			if !fileShouldBeProcessed {
				log.Printf("file %s has not changed, skip processing\n", filePath)
				continue
			}
			log.Printf("processing file %s\n", filePath)
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
			dateLayout := conf.Accounts[bankAccountNameFromFile].DateLayout
			colIndices := accountsToColIndices[bankAccountNameFromFile]
			bankAccountCurrency := conf.Accounts[bankAccountNameFromFile].Currency
			var exchangeRateConfig ExchangeRatesWrapper
			var exchangeRateNum *float64
			if bankAccountCurrency != conf.BaseCurrency {
				rateFilePath := filepath.Join(monthPath, "rate.yaml")
				rateFileData, err := os.ReadFile(rateFilePath)
				if err != nil {
					return fmt.Errorf("could not read rate file at %s: %w", rateFilePath, err)
				}
				if err := yaml.Unmarshal(rateFileData, &exchangeRateConfig); err != nil {
					return fmt.Errorf("error parsing rate file: %w", err)
				}
				for _, rate := range exchangeRateConfig.ExchangeRates {
					if rate.From == bankAccountCurrency {
						exchangeRateNum = &rate.Rate
					}
				}

			}

			// check if headers config at least all exist in file headers. It could
			// be that there are headers in the file that don't exist in config and
			// and that's ok.
			for _, headerInConfig := range headersInConfig {
				if !slices.Contains(headersInFile, headerInConfig) {
					return fmt.Errorf("header '%s' in file: '%s' is not a valid header for this account: Check trackit.yaml", headerInConfig, filePath)
				}
			}

			for _, row := range dataRows {
				date, err := parseDate(dateLayout, row[colIndices["transaction_date"]])
				if date == nil {
					return fmt.Errorf("error parsing date: %v for account %s", row[colIndices["transaction_date"]], bankAccountNameFromFile)
				}
				if err != nil {
					return fmt.Errorf("error parsing date %s: %v", *date, err)
				}
				var amount float64
				thousandsSeparator := conf.Accounts[bankAccountNameFromFile].ThousandsSeparator
				depositIndx, depositIndxExists := colIndices["deposit"]
				withdrawlIndx, withdrawlIndxExists := colIndices["withdrawl"]
				amountIndx, amountIndxExists := colIndices["amount"]
				if amountIndxExists {
					amountStr := row[amountIndx]
					parsedAmount, err := parseAmount(amountStr, thousandsSeparator)
					if err != nil {
						return fmt.Errorf("error parsing amount: %s", amountStr)
					}
					if parsedAmount == nil {
						return fmt.Errorf("parsed amount is nil in: %s", filePath)
					}
					amount = *parsedAmount
				} else {
					if !depositIndxExists || !withdrawlIndxExists {
						return fmt.Errorf("must define a withdrawl and deposit column for: %s", filePath)
					}
					depositStr := row[depositIndx]
					parsedDeposit, err := parseAmount(depositStr, thousandsSeparator)
					if err != nil {
						return fmt.Errorf("error parsing deposit amount %s in %s", depositStr, filePath)
					}
					withdrawlStr := row[withdrawlIndx]
					parsedWithdrawl, err := parseAmount(withdrawlStr, thousandsSeparator)
					if err != nil {
						return fmt.Errorf("error parsing withdrawl amount %s in %s", withdrawlStr, filePath)
					}
					if parsedDeposit == nil {
						return fmt.Errorf("parsed deposit is null in %s", filePath)
					}
					if parsedWithdrawl == nil {
						return fmt.Errorf("parsed withrawl is null in %s", filePath)

					}
					amount = *parsedDeposit - *parsedWithdrawl

				}
				if exchangeRateNum != nil {
					targetAmount := amount * *exchangeRateNum
					roundedAmount := RoundAmount(targetAmount)
					amount = roundedAmount
				}

				transaction := Transaction{Date: *date, Amount: amount, CounterParty: row[colIndices["counter_party"]]}
				var bankAccountId int64
				err = db.QueryRow("SELECT id FROM accounts where name=?", bankAccountNameFromFile).Scan(&bankAccountId)
				if err != nil {
					return fmt.Errorf("error getting bank account ID for %s", bankAccountNameFromFile)
				}
				categoryName, err := getCategory(conf, transaction.CounterParty)
				if err != nil {
					return err
				}
				var categoryId int
				if categoryName != nil {
					err := db.QueryRow("SELECT id FROM categories WHERE name=?", *categoryName).Scan(&categoryId)
					if err != nil {
						return fmt.Errorf("error getting category ID: %w", err)
					}
				}
				_, err = tx.Exec("INSERT INTO transactions (account_id, date, amount, counter_party, category_id) VALUES (?, ?, ?, ?, ?)",
					bankAccountId, transaction.Date, transaction.Amount, transaction.CounterParty, &categoryId)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error inserting transaction: %w", err)
				}
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("error committing transactions to database: %w", err)
			}
		}

	}
	return nil
}

func getCategory(conf *config.Config, counterPayer string) (*string, error) {
	for categoryName, counterPayers := range conf.Categories {
		for _, regexpStr := range counterPayers {
			// @TODO In efficent to compile so many times...
			re, err := regexp.Compile(regexpStr)
			if err != nil {
				return nil, fmt.Errorf("error getting category: %w", err)
			}
			if re.MatchString(counterPayer) {
				return &categoryName, nil
			}
		}
	}
	return nil, nil
}
