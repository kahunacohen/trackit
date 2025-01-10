package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Account struct {
	Currency           string              `yaml:"currency"`
	DateLayout         string              `yaml:"date_layout"`
	Headers            []map[string]string `yaml:"headers"`
	ThousandsSeparator string              `yaml:"thousands_separator"`
}
type Config struct {
	Accounts     map[string]Account  `yaml:"accounts"`
	BaseCurrency string              `yaml:"base_currency"`
	Categories   map[string][]string `yaml:"categories"`
}

func ParseConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read config file at %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	return &config, nil
}

// Returns a map of this form:
// {accountName: {"transaction_date": 0, "counter_party": 3, "amount": 4}}
// This dymamically tells us for each account what column index in the CSV file maps to what
// database table.
func (c *Config) ColIndices() map[string]map[string]int {
	accountToColIndices := make(map[string]map[string]int)
	for accountName, account := range c.Accounts {
		colIndexMap := make(map[string]int)
		for i, headerMap := range account.Headers {
			tableName := headerMap["table"]
			if tableName == "transaction_date" || tableName == "counter_party" || tableName == "amount" || tableName == "deposit" || tableName == "withdrawl" {
				colIndexMap[tableName] = i
			}
		}
		accountToColIndices[accountName] = colIndexMap
	}
	return accountToColIndices
}

func (c *Config) Headers(accountName string) []string {
	headers := c.Accounts[accountName].Headers
	ret := make([]string, len(headers))
	for i, h := range headers {
		ret[i] = h["name"]
	}
	return ret
}
