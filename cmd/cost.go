package cmd

import (
	"fmt"
	"os"

	"costctl/pkg/report"

	"github.com/spf13/cobra"
)

var (
	daysLimit int
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Multi-cloud cost visibility and attribution",
	Long:  `Gather billing and consumption data from configured cloud providers.`,
}

var costListCmd = &cobra.Command{
	Use:   "list",
	Short: "List historical cost records by provider, region, and service",
	Run: func(cmd *cobra.Command, args []string) {
		costs, err := gatherCosts(daysLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		out, cleanup, err := getOutputStream()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		switch Format {
		case "json":
			report.PrintJSON(nil, costs, out)
		case "table":
			report.PrintCostList(costs, out)
		case "html":
			fmt.Fprintln(os.Stderr, "HTML formatting is only supported for the 'waste' command. Falling back to Table.")
			report.PrintCostList(costs, out)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported format: %s. Falling back to Table.\n", Format)
			report.PrintCostList(costs, out)
		}
	},
}

var costSummaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Summarize historical costs and cloud market share",
	Run: func(cmd *cobra.Command, args []string) {
		costs, err := gatherCosts(daysLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		out, cleanup, err := getOutputStream()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		switch Format {
		case "json":
			report.PrintJSON(nil, costs, out)
		case "table":
			report.PrintCostSummary(costs, out)
		case "html":
			fmt.Fprintln(os.Stderr, "HTML formatting is only supported for the 'waste' command. Falling back to Table.")
			report.PrintCostSummary(costs, out)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported format: %s. Falling back to Table.\n", Format)
			report.PrintCostSummary(costs, out)
		}
	},
}

func init() {
	costCmd.PersistentFlags().IntVarP(&daysLimit, "days", "n", 30, "number of historical days of cost data to analyze")
	
	costCmd.AddCommand(costListCmd)
	costCmd.AddCommand(costSummaryCmd)
	rootCmd.AddCommand(costCmd)
}
