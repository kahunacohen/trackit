/*
Copyright © 2024 Aaron Cohen <aaroncohendev@gmail.com>
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

var transactionImportCmd = &cobra.Command{
	Use:   "import",
	Short: "imports transactions",
	Long: `imports transactions by parsing CSV files in the data directory. This will
not parse files whose transactions that already have been added and will ignore non-CSV files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verbose, _ = rootCmd.PersistentFlags().GetBool("verbose")
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
	transactionCmd.AddCommand(transactionImportCmd)
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
	logF(verbose, "walking file path: %s\n", dataPath)
	err = filepath.Walk(dataPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.ToLower(filepath.Ext(path)) == ".csv" {
			fileName := filepath.Base(path)
			if !validateFileName(fileName, conf) {
				return fmt.Errorf("file name '%s' is invalid: it must be a name of a bank account (with spaces separated by '_') defined in trackit.yaml with a .csv extension", path)
			}
			logF(verbose, "found CSV file: %s", path)
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
			hashFromDb, err := txQueries.ReadHashFromFileName(ctx, fileName)
			if err != nil && err != sql.ErrNoRows {
				tx.Rollback()
				return fmt.Errorf("error looking up hash from db for %s: %v", path, err)
			}
			if fileHash == hashFromDb {
				logF(verbose, "file %s has not changed, skip processing\n", path)
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
			accountNameFromFilePtr := getAccountNameFromFileName(conf, fileName)
			if accountNameFromFilePtr == nil {
				return fmt.Errorf("no matching account name in trackit.yaml for file: '%s'", fileName)
			}
			accountNameFromFile := *accountNameFromFilePtr
			accountFromConf := conf.Accounts[accountNameFromFile]
			dataRows := records[1:]
			if len(dataRows) == 0 {
				tx.Rollback()
				return fmt.Errorf("file %s has no records", path)
			}
			headersInConfig := conf.Headers(accountNameFromFile)
			dateLayout := accountFromConf.DateLayout
			colIndices := accountsToColIndices[accountNameFromFile]
			bankAccountCurrency := accountFromConf.Currency

			// Insert bank account name into db if it doesn't exist. @TODO put a unique constraint
			// on account.name then we can use IGNORE in sqlite.
			// _, err = txQueries.ReadAccountIdByName(ctx, accountNameFromFile)
			// if err != nil {
			// 	if err == sql.ErrNoRows {
			// 		// Account doesn't exist so create it.
			// 		if err := txQueries.CreateAccountIfNotExists(ctx, accountNameFromFile); err != nil {
			// 			tx.Rollback()
			// 			return fmt.Errorf("error creating account name %s in db: %w", accountNameFromFile, err)
			// 		}
			// 		logF(verbose, "creating new account in DB: %s", accountNameFromFile)
			// 	} else {
			// 		tx.Rollback()
			// 		return fmt.Errorf("error trying to get account name %s: %w", accountNameFromFile, err)
			// 	}
			// }

			// if err := txQueries.CreateAccountIfNotExists(ctx, accountNameFromFile); err != nil {
			// 	tx.Rollback()
			// 	return fmt.Errorf("error creating account name %s in db: %w", accountNameFromFile, err)
			// }
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
					return fmt.Errorf("error parsing date: %v with layout: %s for account %s", row[colIndices["transaction_date"]], dateLayout, accountNameFromFile)
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
								return fmt.Errorf(`no rate defined from %s to %s for month: %s, file: %s Create currency
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

				// Get bank account if it exists, otherwise create it in DB. @TODO create function
				bankAccountId, err := txQueries.ReadAccountIdByName(ctx, accountNameFromFile)
				if err != nil {
					if err == sql.ErrNoRows {
						_, err = txQueries.CreateAccount(ctx, models.CreateAccountParams{Name: accountNameFromFile, Currency: bankAccountCurrency})
						if err != nil {
							tx.Rollback()
							return fmt.Errorf("error creating account in DB: %s: %w", accountNameFromFile, err)
						}
						// Need to do this because we are in the middle of a transaction.
						// @TODO one solution is to create any necessary bank acccounts first before this loop by
						// looking at the config and creating then as a pre-step.
						err = tx.QueryRowContext(ctx, "SELECT last_insert_rowid()").Scan(&bankAccountId)
						logF(verbose, "creating account for %s with ID: %d", accountNameFromFile, bankAccountId)

						if err != nil {
							tx.Rollback()
							return fmt.Errorf("error getting just inserted bankAccountID: %w", err)
						}
					} else {
						tx.Rollback()
						return fmt.Errorf("error getting bank account ID for %s: %w", accountNameFromFile, err)
					}
				}
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("error getting bank account ID for %s: %w", accountNameFromFile, err)
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
				if accountFromConf.DebitAsPositive {
					amount = -amount
				}
				logF(verbose, "inserting transaction for %f, in account: %s\n", amount, accountNameFromFile)
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
				logF(verbose, "file %s had never been processed, insert hash to db\n", path)
				if err := txQueries.CreateFile(ctx, models.CreateFileParams{Name: fileName, Hash: fileHash}); err != nil {
					tx.Rollback()
					return fmt.Errorf("error inserting initial file hash: %w", err)
				}
			} else {
				logF(verbose, "file %s changed, update hash\n", path)
				err = txQueries.UpdateFileHashByName(ctx, models.UpdateFileHashByNameParams{Hash: fileHash, Name: fileName})
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

func getAccountNameFromFileName(conf *config.Config, fileName string) *string {
	for k := range conf.Accounts {
		if strings.Contains(fileName, k) {
			return &k
		}
	}
	return nil
}

func validateFileName(fileName string, conf *config.Config) bool {
	for accountName := range conf.Accounts {
		if strings.Contains(fileName, accountName) {
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
