package cmd

import (
	"fmt"
	"os"

	"costctl/pkg/config"

	"github.com/spf13/cobra"
)

var (
	CfgFile    string
	Format     string
	OutputFile string
	DemoMode   bool
	Cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "costctl",
	Short: "costctl is a multi-cloud cost visibility and waste optimization CLI tool",
	Long: `costctl parses billing data and scans live resources across AWS, Azure, and GCP 
to help DevOps teams identify cost attribution and waste optimization options.

Running without options shows help. Use costctl --demo waste find to see a dry run.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&CfgFile, "config", "c", "", "config file (default is $HOME/.costctl.yaml or ./.costctl.yaml)")
	rootCmd.PersistentFlags().StringVarP(&Format, "format", "f", "table", "output format (table, json, html)")
	rootCmd.PersistentFlags().StringVarP(&OutputFile, "output", "o", "", "output file path (default is stdout)")
	rootCmd.PersistentFlags().BoolVarP(&DemoMode, "demo", "d", false, "run in demo/mock mode with realistic sample data")
}

func initConfig() {
	var err error
	Cfg, err = config.LoadConfig(CfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
}
