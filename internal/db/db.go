package db

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
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
	"github.com/kahunacohen/trackit/internal/models"
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

func GetCachedDbPath() (*string, error) {
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

//go:embed schema.sql
var schemaSQL string

func InitSchema(conf *config.Config, db *sql.DB) error {
	if _, err := db.Exec(schemaSQL); err != nil {
		return err
	}
	return nil
}

type Transaction struct {
	Id           int64
	Amount       float64
	Category     *string
	Hash         string
	CounterParty string
	Date         string
}

func validateDateDirectoryName(name string) bool {
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
		if accountName+".csv" == fileName {
			return true
		}
	}
	return false
}
func GetAccountTransactions(db *sql.DB, accountName string, date string) ([]models.ReadTransactionsRow, error) {
	var rows []models.ReadTransactionsRow
	var err error
	queries := models.New(db)
	// account and date are not set
	ctx := context.Background()

	// @TODO this is a bit messy, repeated code, etc. Maybe make a wrapper function
	// that handles the distinct types but with same fields.
	if accountName == "" && date == "" {
		rows, err = queries.ReadTransactions(ctx)
		if err != nil {
			return nil, err
		}
	} else if accountName != "" && date != "" {
		xs, err := queries.ReadTransactionsByAccountNameAndDate(ctx, models.ReadTransactionsByAccountNameAndDateParams{
			AccountName: accountName,
			Date:        date})
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
		// account name is set but not date
	} else if accountName != "" && date == "" {
		xs, err := queries.ReadTransactionsByAccountName(ctx, accountName)
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
		// date is set but not account
	} else {
		xs, err := queries.ReadTransactionsByDate(ctx, date)
		if err != nil {
			return nil, err
		}
		// convert type with same shape to the general type since they have the same fields.
		for _, x := range xs {
			rows = append(rows, models.ReadTransactionsRow(x))
		}
	}
	if err != nil {
		return nil, err
	}
	return rows, nil
}
func GetCategoryAggregation(db *sql.DB, account string, date string) ([]models.AggregateTransactionsRow, error) {
	queries := models.New(db)
	ctx := context.Background()
	var err error
	var rows []models.AggregateTransactionsRow
	if account == "" && date == "" {
		rows, err = queries.AggregateTransactions(ctx)
		if err != nil {
			return nil, fmt.Errorf("error aggregating rows: %w", err)
		}
	} else if account != "" && date == "" {
		xs, err := queries.AggregateTransactionsByAccountName(ctx, account)
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	} else if account == "" && date != "" {
		xs, err := queries.AggregateTransactionsByDate(ctx, date)
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	} else {
		xs, err := queries.AggregateTransactionsByAccountNameAndDate(ctx,
			models.AggregateTransactionsByAccountNameAndDateParams{AccountName: account, Date: date})
		if err != nil {
			return nil, fmt.Errorf("error aggreating rows: %w", err)
		}
		for _, x := range xs {
			rows = append(rows, models.AggregateTransactionsRow(x))
		}
	}
	return rows, nil
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

func ProcessFiles(conf *config.Config, db *sql.DB) error {
	// First create a map of account name to db table names to indices
	// like: {bank_of_america: {date: 0}} etc.
	// so we can know each bank account's csv structure.
	queries := models.New(db)
	ctx := context.Background()
	accountsToColIndices := conf.ColIndices()
	dateEntries, err := os.ReadDir(conf.Data)
	if err != nil {
		return err
	}
	for _, dateEntry := range dateEntries {
		dateName := dateEntry.Name()
		validName := validateDateDirectoryName(dateName)
		if !validName {
			log.Printf("skipping directory '%s'. Not a valid month directory in the form YYYY-mm", dateName)
			continue
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
			var isFileHashInDB bool = true
			hashFromDb, err := queries.ReadHashFromFileName(ctx, filePath)
			if err != nil {
				if err == sql.ErrNoRows {
					// The file never had a hash and has never been seen before.
					log.Printf("file %s has never been processed, insert hash to db\n", filePath)
					if err := queries.CreateFile(ctx, models.CreateFileParams{Name: filePath, Hash: fileHash}); err != nil {
						return fmt.Errorf("error inserting initial file hash: %w", err)
					}

					// File should be processed if hash not in db at all.
					isFileHashInDB = false
				} else {
					// Some other error trying to get hash.
					return fmt.Errorf("error looking up hash from db for %s: %v", filePath, err)
				}
			}
			if isFileHashInDB || fileHash == hashFromDb {
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
			bankAccountNameFromFile := strings.Replace(fileName, ".csv", "", 1)
			accountFromConf := conf.Accounts[bankAccountNameFromFile]
			dataRows := records[1:]
			if len(dataRows) == 0 {
				return fmt.Errorf("file %s has no records", filePath)
			}
			headersInConfig := conf.Headers(bankAccountNameFromFile)
			dateLayout := accountFromConf.DateLayout
			colIndices := accountsToColIndices[bankAccountNameFromFile]
			bankAccountCurrency := accountFromConf.Currency
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
				date, err := time.Parse(dateLayout, row[colIndices["transaction_date"]])
				if err != nil {
					return fmt.Errorf("error parsing date: %v for account %s", row[colIndices["transaction_date"]], bankAccountNameFromFile)
				}
				if err != nil {
					return fmt.Errorf("error parsing date %s: %v", date, err)
				}
				var amount float64
				thousandsSeparator := accountFromConf.ThousandsSeparator
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
				counterParty := row[colIndices["counter_party"]]
				bankAccountId, err := queries.ReadAccountIdByName(ctx, bankAccountNameFromFile)
				if err != nil {
					return fmt.Errorf("error getting bank account ID for %s", bankAccountNameFromFile)
				}
				categoryName, err := getCategory(conf, counterParty)
				if err != nil {
					return err
				}
				var categoryId int64
				if categoryName != nil {
					categoryId, err = queries.ReadCategoryIdByName(ctx, *categoryName)
					if err != nil {
						return fmt.Errorf("error getting category ID: %w", err)
					}
				}
				err = queries.CreateTransaction(ctx, models.CreateTransactionParams{
					AccountID:    bankAccountId,
					Date:         date.Format("2006-01-02"),
					Amount:       amount,
					CounterParty: counterParty,
					CategoryID:   ToNullInt64(&categoryId)})
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error inserting transaction: %w", err)
				}
			}
			err = queries.UpdateFileHashByName(ctx, models.UpdateFileHashByNameParams{Hash: fileHash, Name: filePath})
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error updating file hash for file: %s: %w", filePath, err)
			}
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("error committing transactions to database: %w", err)
			}
		}

	}
	return nil
}

func ToNullInt64(val *int64) sql.NullInt64 {
	if val != nil {
		return sql.NullInt64{Int64: *val, Valid: true}
	} else {
		return sql.NullInt64{Valid: false}
	}
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
