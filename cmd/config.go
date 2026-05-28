package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"costctl/pkg/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage costctl settings and thresholds",
	Long:  `Create, edit, or view active configurations, including normalized tag schemas and idle resource metrics.`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default .costctl.yaml configuration file in the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		path := ".costctl.yaml"
		if _, err := os.Stat(path); err == nil {
			fmt.Fprintf(os.Stderr, "Error: File %s already exists in the current directory.\n", path)
			os.Exit(1)
		}

		defaultCfg := config.DefaultConfig()
		data, err := yaml.Marshal(defaultCfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling default configuration: %v\n", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile(path, data, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing configuration file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully created default configuration file: %s\n", path)
		fmt.Println("You can now customize tag mappings and waste thresholds inside this file.")
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the currently active configuration settings",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := yaml.Marshal(Cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error marshalling active configuration: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("--- Active Configuration ---")
		fmt.Println(string(data))
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	rootCmd.AddCommand(configCmd)
}
