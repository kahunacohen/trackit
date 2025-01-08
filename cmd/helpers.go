package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
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
