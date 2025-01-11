/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/kahunacohen/trackit/internal/config"
	"github.com/kahunacohen/trackit/internal/models"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "imports transactions",
	Long: `imports transactions by parsing CSV files in the data directory. This will
not parse files whose transactions that already have been added.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		db, err := getDB()
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}
		defer db.Close()
		queries := models.New(db)
		configPath, err := queries.ReadSettingByName(context.Background(), "config-file")
		if err != nil {
			return fmt.Errorf("error getting config-file path from db: %w", err)
		}

		conf, err := config.ParseConfig(configPath)
		if err != nil {
			return err
		}
		err = processFiles(conf, db)
		if err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}

func processFiles(conf *config.Config, db *sql.DB) error {
	// First create a map of account name to db table names to indices
	// like: {bank_of_america: {date: 0}} etc.
	// so we can know each bank account's csv structure.
	queries := models.New(db)
	ctx := context.Background()
	accountsToColIndices := conf.ColIndices()
	dataPath, err := queries.ReadSettingByName(ctx, "data-path")
	if err != nil {
		return fmt.Errorf("error getting data path setting: %w", err)
	}
	dataPath, err = filepath.Abs(dataPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for data directory: %w", err)
	}
	dateEntries, err := os.ReadDir(dataPath)
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
		monthPath := filepath.Join(dataPath, dateName)
		fileEntries, err := os.ReadDir(monthPath)
		if err != nil {
			return err
		}
		if len(fileEntries) == 0 {
			return fmt.Errorf("month directory '%s' contains no CSV files", dateName)
		}
		for _, fileEntry := range fileEntries {
			fileName := fileEntry.Name()
			if fileName == "rates.yaml" || fileName == "rates.yml" {
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
				rateFilePath := filepath.Join(monthPath, "rates.yaml")
				rateFileData, err := os.ReadFile(rateFilePath)
				if err != nil {
					return fmt.Errorf("could not get a conversion rate from a rates.yaml file at %s. Error: %w", rateFilePath, err)
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
					roundedAmount := roundAmount(targetAmount)
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
					CategoryID:   toNullInt64(&categoryId)})
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

func validateDateDirectoryName(name string) bool {
	split := strings.Split(name, "-")
	if len(split) != 2 {
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

func validateFileName(fileName string, conf *config.Config) bool {
	for accountName := range conf.Accounts {
		if accountName+".csv" == fileName {
			return true
		}
	}
	return false
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

type ExchangeRate struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

type ExchangeRatesWrapper struct {
	ExchangeRates []ExchangeRate `json:"exchange_rates"`
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

func toNullInt64(val *int64) sql.NullInt64 {
	if val != nil {
		return sql.NullInt64{Int64: *val, Valid: true}
	} else {
		return sql.NullInt64{Valid: false}
	}
}
