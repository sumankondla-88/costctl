package cmd

import (
	"fmt"
	"os"

	"costctl/pkg/analyzer"
	"costctl/pkg/report"

	"github.com/spf13/cobra"
)

var (
	failOnWaste bool
	maxWasteVal float64
)

var wasteCmd = &cobra.Command{
	Use:   "waste",
	Short: "Detect idle and unattached cloud resources",
	Long:  `Scan your cloud environments for underutilized, orphaned, or stale resources to detect cost waste.`,
}

var wasteFindCmd = &cobra.Command{
	Use:   "find",
	Short: "Find wasteful resources and show savings recommendations",
	Run: func(cmd *cobra.Command, args []string) {
		resources, err := gatherResources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result := analyzer.AnalyzeResources(resources, Cfg)

		out, cleanup, err := getOutputStream()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		defer cleanup()

		// If HTML and output file is not specified, default to a file named 'report.html'
		if Format == "html" && OutputFile == "" {
			OutputFile = "report.html"
			cleanup() // Close current stdout writer wrapper
			
			// Re-initialize writer pointing to the file
			out, cleanup, err = getOutputStream()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer cleanup()
			
			fmt.Fprintf(os.Stderr, "Generating HTML dashboard report: %s\n", OutputFile)
		}

		switch Format {
		case "json":
			report.PrintJSON(result, nil, out)
		case "table":
			report.PrintWasteFindings(result, out)
		case "html":
			err := report.RenderHTML(result, resources, out)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating HTML: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Fprintf(os.Stderr, "Unsupported format: %s. Falling back to Table.\n", Format)
			report.PrintWasteFindings(result, out)
		}

		// Perform CI/CD waste limit enforcement if enabled
		handleWasteLimits(result)
	},
}

var wasteAnalyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Perform a waste audit, return summary findings, and check budget compliance",
	Run: func(cmd *cobra.Command, args []string) {
		resources, err := gatherResources()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		result := analyzer.AnalyzeResources(resources, Cfg)

		fmt.Printf("Multi-Cloud FinOps Audit Report:\n")
		fmt.Printf("---------------------------------\n")
		fmt.Printf("Total Scanned Resource Cost (Monthly estimate): $%.2f\n", result.TotalMonthlyCost)
		fmt.Printf("Potential Waste Savings (Monthly estimate):      $%.2f\n", result.TotalSavingsPotential)
		fmt.Printf("Waste Share percentage:                        %.1f%%\n", result.WastePercent)
		fmt.Printf("Detected Waste Asset count:                    %d\n", len(result.Findings))

		handleWasteLimits(result)
	},
}

func handleWasteLimits(result *analyzer.AnalysisResult) {
	if !failOnWaste {
		return
	}

	threshold := Cfg.Budget.FailOnWasteThreshold
	if maxWasteVal >= 0 {
		threshold = maxWasteVal
	}

	if result.TotalSavingsPotential > threshold {
		fmt.Fprintf(os.Stderr, "\n[FAIL] FinOps CI/CD Policy Check: Potential monthly waste ($%.2f) exceeds limit ($%.2f)\n", 
			result.TotalSavingsPotential, threshold)
		os.Exit(1)
	}

	fmt.Printf("\n[SUCCESS] FinOps CI/CD Policy Check: Waste ($%.2f) is within limit ($%.2f)\n", 
		result.TotalSavingsPotential, threshold)
}

func init() {
	wasteCmd.PersistentFlags().BoolVar(&failOnWaste, "fail-on-waste", false, "fail the CLI execution with exit code 1 if potential waste exceeds thresholds (useful in CI/CD)")
	wasteCmd.PersistentFlags().Float64Var(&maxWasteVal, "max-waste", -1.0, "override config value for maximum allowable monthly waste before failing (USD)")

	wasteCmd.AddCommand(wasteFindCmd)
	wasteCmd.AddCommand(wasteAnalyzeCmd)
	rootCmd.AddCommand(wasteCmd)
}
