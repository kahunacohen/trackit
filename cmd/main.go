package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var rootCmd = &cobra.Command{
	Use:   "trackit",
	Short: "Trackit is a CLI tool for tracking tasks.",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the configuration from a YAML file",
	Run:   initRun,
}

var configFile string

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "Path to the YAML configuration file")
}

func initRun(cmd *cobra.Command, args []string) {
	// Logic to read and process the YAML file
	if err := setupConfig(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting up config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Configuration initialized successfully!")
}

func setupConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Do something with the config
	fmt.Printf("Loaded config: %+v\n", config)
	return nil
}

func main() {
	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
