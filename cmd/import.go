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
	"io/fs"
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
)

var verbose bool

var importCmd = &cobra.Command{
	Use:   "import",
	Short: "imports transactions",
	Long: `imports transactions by parsing CSV files in the data directory. This will
not parse files whose transactions that already have been added and will ignore non-CSV files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		v, _ := cmd.Flags().GetBool("verbose")
		verbose = v
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

type rateCacheKey struct {
	Date       string
	ToCurrency string
}

var exchangeRateCache = make(map[rateCacheKey]float64)

func processFiles(conf *config.Config, db *sql.DB) error {
	dbQueries := models.New(db)
	ctx := context.Background()
	accountsToColIndices := conf.AccountColumnIndices()
	dataPath, err := dbQueries.ReadSettingByName(ctx, "data-dir")
	if err != nil {
		return fmt.Errorf("error getting data directory: %w", err)
	}
	dataPath, err = filepath.Abs(dataPath)
	if err != nil {
		return fmt.Errorf("error getting absolute path for data directory: %w", err)
	}
	logF("walking file path: %s\n", verbose, dataPath)
	err = filepath.Walk(dataPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.ToLower(filepath.Ext(path)) == ".csv" {
			fileName := filepath.Base(path)
			if !validateFileName(fileName, conf) {
				return fmt.Errorf("file name '%s' is invalid: it must be a name of a bank account (with spaces separated by '_') defined in trackit.yaml with a .csv extension", path)
			}
			logF("found CSV file: %s", verbose, path)
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("error opening %s: %w", path, err)
			}
			fileHash, err := computeFileHash(file)
			if err != nil {
				return fmt.Errorf("problem hashing file: %w", err)
			}

			logLn("begin transaction", verbose)

			tx, err := db.Begin()
			if err != nil {
				return fmt.Errorf("error beginning db transaction when inserting transactions: %w", err)
			}
			txQueries := models.New(tx)
			hashFromDb, err := txQueries.ReadHashFromFileName(ctx, path)
			if err != nil && err != sql.ErrNoRows {
				tx.Rollback()
				return fmt.Errorf("error looking up hash from db for %s: %v", path, err)
			}
			if fileHash == hashFromDb {
				logF("file %s has not changed, skip processing\n", verbose, path)
				tx.Rollback()
				return nil
			}
			reader := csv.NewReader(file)
			records, err := reader.ReadAll()
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error reading %s: %w", path, err)
			}
			file.Close()
			if len(records) < 2 {
				tx.Rollback()
				return fmt.Errorf("there are less than 2 rows for file: %s", path)
			}
			headersInFile := records[0]
			bankAccountNameFromFile := strings.Replace(fileName, ".csv", "", 1)
			accountFromConf := conf.Accounts[bankAccountNameFromFile]
			dataRows := records[1:]
			if len(dataRows) == 0 {
				tx.Rollback()
				return fmt.Errorf("file %s has no records", path)
			}
			headersInConfig := conf.Headers(bankAccountNameFromFile)
			dateLayout := accountFromConf.DateLayout
			colIndices := accountsToColIndices[bankAccountNameFromFile]
			bankAccountCurrency := accountFromConf.Currency
			for _, headerInConfig := range headersInConfig {
				if !slices.Contains(headersInFile, headerInConfig) {
					tx.Rollback()
					return fmt.Errorf("header '%s' in file: '%s' is not a valid header for this account: Check trackit.yaml", headerInConfig, path)
				}
			}
			for _, row := range dataRows {
				rowDateStr := row[colIndices["transaction_date"]]
				date, err := time.Parse(dateLayout, rowDateStr)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error parsing date: %v for account %s", row[colIndices["transaction_date"]], bankAccountNameFromFile)
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
						tx.Rollback()
						return fmt.Errorf("error parsing amount: %s", amountStr)
					}
					if parsedAmount == nil {
						tx.Rollback()
						return fmt.Errorf("parsed amount is nil in: %s", path)
					}
					amount = *parsedAmount
				} else {
					if !depositIndxExists || !withdrawlIndxExists {
						tx.Rollback()
						return fmt.Errorf("must define a withdrawl and deposit column for: %s", path)
					}
					depositStr := row[depositIndx]
					parsedDeposit, err := parseAmount(depositStr, thousandsSeparator)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("error parsing deposit amount %s in %s", depositStr, path)
					}
					withdrawlStr := row[withdrawlIndx]
					parsedWithdrawl, err := parseAmount(withdrawlStr, thousandsSeparator)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("error parsing withdrawl amount %s in %s", withdrawlStr, path)
					}
					if parsedDeposit == nil {
						tx.Rollback()
						return fmt.Errorf("parsed deposit is null in %s", path)
					}
					if parsedWithdrawl == nil {
						tx.Rollback()
						return fmt.Errorf("parsed withrawl is null in %s", path)

					}
					amount = *parsedDeposit - *parsedWithdrawl
				}
				if bankAccountCurrency != conf.BaseCurrency {
					normalizedTransactionDate := date.Format("2006-01")
					cacheKey := rateCacheKey{Date: normalizedTransactionDate, ToCurrency: bankAccountCurrency}
					rate, ok := exchangeRateCache[cacheKey]
					if !ok {
						rate, err = txQueries.ReadRateFromSymbols(ctx, models.ReadRateFromSymbolsParams{
							Fromsymbol: bankAccountCurrency,
							Month:      normalizedTransactionDate})
						if err != nil {
							if err == sql.ErrNoRows {
								tx.Rollback()
								return fmt.Errorf(`no rate defined from %s to %s for month: %s, file: %s Create curency
(trackit currency create) and rate (trackit rate create) to define a conversion rate for this month`, bankAccountCurrency, conf.BaseCurrency, normalizedTransactionDate, path)
							} else {
								tx.Rollback()
								return fmt.Errorf("error reading rate %s to %s for month %s from DB: %w", bankAccountCurrency, conf.BaseCurrency, normalizedTransactionDate, err)
							}
						}
						exchangeRateCache[cacheKey] = rate
					}
					targetAmount := amount * rate
					roundedAmount := roundAmount(targetAmount)
					amount = roundedAmount
				}

				counterParty := row[colIndices["counter_party"]]
				bankAccountId, err := txQueries.ReadAccountIdByName(ctx, bankAccountNameFromFile)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error getting bank account ID for %s", bankAccountNameFromFile)
				}
				categoryName, err := getCategory(conf, counterParty)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error getting category: %w", err)
				}
				var categoryId int64
				if categoryName != nil {
					categoryId, err = txQueries.ReadCategoryIdByName(ctx, *categoryName)
					if err != nil {
						tx.Rollback()
						return fmt.Errorf("error getting category ID: %w", err)
					}
				}
				logF("inserting transaction for %f\n", verbose, amount)
				err = txQueries.CreateTransaction(ctx, models.CreateTransactionParams{
					AccountID:    sql.NullInt64{Valid: true, Int64: bankAccountId},
					Date:         date.Format("2006-01-02"),
					Amount:       amount,
					CounterParty: counterParty,
					CategoryID:   toNullInt64(&categoryId)})
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error inserting transaction: %w", err)
				}
			} // end iteration of data rows in file
			if hashFromDb == "" {
				logF("file %s had never been processed, insert hash to db\n", verbose, path)
				if err := txQueries.CreateFile(ctx, models.CreateFileParams{Name: path, Hash: fileHash}); err != nil {
					tx.Rollback()
					return fmt.Errorf("error inserting initial file hash: %w", err)
				}
			} else {
				logF("file %s changed, update hash\n", verbose, path)
				err = txQueries.UpdateFileHashByName(ctx, models.UpdateFileHashByNameParams{Hash: fileHash, Name: path})
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error updating file hash for file: %s: %w", path, err)
				}
			}
			logLn("commit db transaction", verbose)
			err = tx.Commit()
			if err != nil {
				return fmt.Errorf("error committing transactions to database: %w", err)
			}
		} // end if csv
		return nil
	}) // end of walk
	if err != nil {
		return err
	}
	return nil
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

func getCategory(conf *config.Config, counterParty string) (*string, error) {
	for categoryName, counterParties := range conf.Categories {
		for _, regexpStr := range counterParties {
			// @TODO In efficent to compile so many times...
			re, err := regexp.Compile(regexpStr)
			if err != nil {
				return nil, fmt.Errorf("error getting category: %w", err)
			}
			if re.MatchString(counterParty) {
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
