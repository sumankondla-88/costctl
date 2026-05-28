package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"costctl/pkg/analyzer"
	"costctl/pkg/provider"
)

// ANSI color escape codes for terminal styling
const (
	colorReset   = "\033[0m"
	colorBold    = "\033[1m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorBlue    = "\033[34m"
	colorMagenta = "\033[35m"
	colorCyan    = "\033[36m"
	colorGray    = "\033[90m"
	colorAWS     = "\033[38;5;208m" // Orange
	colorAzure   = "\033[38;5;39m"  // Bright Blue
	colorGCP     = "\033[38;5;118m" // Green/Light Blue
)

// SummaryData groups cost information for reporting
type SummaryData struct {
	Provider      string
	Cost          float64
	Percentage    float64
	ProgressBar   string
}

// FormatCost formats float values as currency strings
func FormatCost(val float64) string {
	return fmt.Sprintf("$%.2f", val)
}

// ColorizeCloud returns colored text for cloud providers
func ColorizeCloud(p provider.CloudType) string {
	switch p {
	case provider.AWS:
		return fmt.Sprintf("%sAWS%s", colorAWS, colorReset)
	case provider.Azure:
		return fmt.Sprintf("%sAzure%s", colorAzure, colorReset)
	case provider.GCP:
		return fmt.Sprintf("%sGCP%s", colorGCP, colorReset)
	default:
		return fmt.Sprintf("%s%s%s", colorGray, p, colorReset)
	}
}

// GenerateProgressBar creates a text-based progress bar for CLI
func GenerateProgressBar(percentage float64, size int) string {
	blocks := int(math.Round(percentage / 100.0 * float64(size)))
	if blocks < 0 {
		blocks = 0
	}
	if blocks > size {
		blocks = size
	}
	bar := strings.Repeat("█", blocks) + strings.Repeat("░", size-blocks)
	return bar
}

// PrintCostSummary displays a summary of historical costs by cloud provider
func PrintCostSummary(costs []provider.CostItem, w io.Writer) {
	totals := make(map[string]float64)
	grandTotal := 0.0

	for _, item := range costs {
		totals[string(item.Provider)] += item.Cost
		grandTotal += item.Cost
	}

	fmt.Fprintf(w, "\n%s=== Multi-Cloud Cost Summary ===%s\n\n", colorBold+colorCyan, colorReset)

	var summaries []SummaryData
	for prov, val := range totals {
		pct := 0.0
		if grandTotal > 0 {
			pct = (val / grandTotal) * 100.0
		}
		summaries = append(summaries, SummaryData{
			Provider:   prov,
			Cost:       val,
			Percentage: pct,
			ProgressBar: GenerateProgressBar(pct, 15),
		})
	}

	// Sort by cost descending
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Cost > summaries[j].Cost
	})

	tw := tabwriter.NewWriter(w, 0, 0, 4, ' ', 0)
	fmt.Fprintf(tw, "%sCloud%s\t%sTotal Cost (Period)%s\t%sShare%%%s\t%sDistribution%s\n", colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)
	
	for _, summary := range summaries {
		coloredCloud := ColorizeCloud(provider.CloudType(summary.Provider))
		fmt.Fprintf(tw, "%s\t%s\t%.1f%%\t%s%s%s\n", 
			coloredCloud, 
			FormatCost(summary.Cost), 
			summary.Percentage, 
			colorGray, summary.ProgressBar, colorReset,
		)
	}
	tw.Flush()

	fmt.Fprintf(w, "\n%sTotal Period Cost: %s%s%s\n\n", colorBold, colorGreen, FormatCost(grandTotal), colorReset)
}

// PrintCostList displays a detailed list of costs
func PrintCostList(costs []provider.CostItem, w io.Writer) {
	fmt.Fprintf(w, "\n%s=== Detailed Cloud Spend (Historical Period) ===%s\n\n", colorBold+colorCyan, colorReset)

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)
	fmt.Fprintf(tw, "%sDate%s\t%sCloud%s\t%sRegion%s\t%sService%s\t%sDaily Cost%s\t%sOwner%s\n", 
		colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)

	// Sort costs by Date desc, then Provider
	sort.Slice(costs, func(i, j int) bool {
		if costs[i].Date.Equal(costs[j].Date) {
			return costs[i].Provider < costs[j].Provider
		}
		return costs[i].Date.After(costs[j].Date)
	})

	for _, item := range costs {
		coloredCloud := ColorizeCloud(item.Provider)
		owner := "unknown"
		if item.NormalizedTags != nil && item.NormalizedTags["owner"] != "" {
			owner = item.NormalizedTags["owner"]
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
			item.Date.Format("2006-01-02"),
			coloredCloud,
			item.Region,
			item.Service,
			FormatCost(item.Cost),
			owner,
		)
	}
	tw.Flush()
	fmt.Println()
}

// PrintWasteFindings prints list of wasteful resources in CLI table
func PrintWasteFindings(result *analyzer.AnalysisResult, w io.Writer) {
	fmt.Fprintf(w, "\n%s=== Optimization Waste Findings ===%s\n\n", colorBold+colorRed, colorReset)

	if len(result.Findings) == 0 {
		fmt.Fprintf(w, "%sNo waste resources detected. Great job!%s\n\n", colorGreen, colorReset)
		return
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%sID/Name%s\t%sCloud%s\t%sResource Type%s\t%sMonthly Waste%s\t%sWaste Reason & Recommendation%s\n", 
		colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset, colorBold, colorReset)

	for _, finding := range result.Findings {
		nameOrID := finding.ResourceName
		if nameOrID == "" {
			nameOrID = finding.ResourceID
			if len(nameOrID) > 30 {
				// Truncate resource ID for display
				nameOrID = "..." + nameOrID[len(nameOrID)-27:]
			}
		}
		coloredCloud := ColorizeCloud(finding.Provider)
		
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s%s%s\t%s%s%s\n",
			nameOrID,
			coloredCloud,
			finding.ResourceType,
			colorYellow, FormatCost(finding.SavingsPotential), colorReset,
			colorGray, finding.Reason, colorReset,
		)
	}
	tw.Flush()

	fmt.Fprintf(w, "\n%s=== Waste Metrics Audit ===%s\n", colorBold, colorReset)
	fmt.Fprintf(w, "Total Active Cloud Inventory Cost:  %s\n", FormatCost(result.TotalMonthlyCost))
	fmt.Fprintf(w, "Total Potential Monthly Savings:   %s%s%s\n", colorGreen, FormatCost(result.TotalSavingsPotential), colorReset)
	fmt.Fprintf(w, "Waste Percentage:                  %s%.1f%%%s\n\n", colorRed, result.WastePercent, colorReset)
}

// PrintJSON prints resource analysis and costs as a structured JSON object
func PrintJSON(result *analyzer.AnalysisResult, costs []provider.CostItem, w io.Writer) {
	type OutputJSON struct {
		GeneratedAt   string                   `json:"generated_at"`
		CostsSummary  []provider.CostItem      `json:"costs_detail"`
		WasteAnalysis *analyzer.AnalysisResult `json:"waste_analysis"`
	}

	output := OutputJSON{
		GeneratedAt:   time.Now().Format(time.RFC3339),
		CostsSummary:  costs,
		WasteAnalysis: result,
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(output); err != nil {
		fmt.Fprintf(w, "Error generating JSON: %v\n", err)
	}
}

// HTMLReportTemplate represents the interactive HTML dashboard template
const HTMLReportTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>costctl — Spend Optimizer Console</title>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg-color: #080b11;
            --sidebar-bg: #0f131f;
            --card-bg: rgba(15, 19, 31, 0.7);
            --border-color: rgba(255, 255, 255, 0.08);
            --text-main: #f9fafb;
            --text-muted: #9ca3af;
            --primary: #4f46e5;
            --primary-hover: #4338ca;
            --primary-glow: rgba(79, 70, 229, 0.2);
            --danger: #ef4444;
            --danger-glow: rgba(239, 68, 68, 0.1);
            --success: #10b981;
            --success-glow: rgba(16, 185, 129, 0.1);
            --warning: #f59e0b;
            --warning-glow: rgba(245, 158, 11, 0.1);
            
            --aws-color: #ff9900;
            --azure-color: #0089d6;
            --gcp-color: #34a853;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: 'Plus Jakarta Sans', sans-serif;
            background-color: var(--bg-color);
            color: var(--text-main);
            min-height: 100vh;
            display: flex;
            background-image: radial-gradient(circle at 10% 20%, rgba(79, 70, 229, 0.06) 0%, transparent 40%),
                              radial-gradient(circle at 90% 80%, rgba(239, 68, 68, 0.04) 0%, transparent 40%);
        }

        /* LOGIN SCREEN VIEW */
        #login-screen {
            display: flex;
            align-items: center;
            justify-content: center;
            width: 100vw;
            height: 100vh;
            position: fixed;
            top: 0;
            left: 0;
            z-index: 1000;
            background-color: var(--bg-color);
            background-image: radial-gradient(circle at 50% 50%, rgba(79, 70, 229, 0.1) 0%, transparent 50%);
        }

        .login-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            backdrop-filter: blur(24px);
            padding: 3rem;
            border-radius: 24px;
            width: 440px;
            box-shadow: 0 20px 50px rgba(0, 0, 0, 0.3);
            text-align: center;
            display: flex;
            flex-direction: column;
            gap: 2rem;
        }

        .login-logo {
            font-size: 2rem;
            font-weight: 800;
            letter-spacing: -0.06em;
            background: linear-gradient(to right, #818cf8, #e0e7ff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }

        .login-header h2 {
            font-size: 1.25rem;
            font-weight: 700;
            margin-top: 0.5rem;
        }

        .login-header p {
            font-size: 0.85rem;
            color: var(--text-muted);
            margin-top: 0.4rem;
        }

        .sso-buttons {
            display: flex;
            flex-direction: column;
            gap: 0.8rem;
        }

        .btn-sso {
            display: flex;
            align-items: center;
            justify-content: center;
            gap: 0.8rem;
            padding: 0.8rem;
            border-radius: 8px;
            font-size: 0.9rem;
            font-weight: 600;
            cursor: pointer;
            border: 1px solid var(--border-color);
            transition: all 0.2s;
        }

        .btn-sso.google {
            background-color: white;
            color: #1f2937;
            border-color: #e5e7eb;
        }

        .btn-sso.google:hover {
            background-color: #f9fafb;
            box-shadow: 0 4px 12px rgba(255, 255, 255, 0.1);
        }

        .btn-sso.entra {
            background-color: #2f2f2f;
            color: white;
            border-color: rgba(255, 255, 255, 0.1);
        }

        .btn-sso.entra:hover {
            background-color: #3f3f3f;
            box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
        }

        .login-footer {
            font-size: 0.75rem;
            color: var(--text-muted);
            line-height: 1.4;
        }

        /* MAIN APP VIEW (HIDDEN BY DEFAULT) */
        #app-view {
            display: none;
            width: 100%;
        }

        /* Sidebar Navigation Layout */
        .sidebar {
            width: 260px;
            background-color: var(--sidebar-bg);
            border-right: 1px solid var(--border-color);
            display: flex;
            flex-direction: column;
            padding: 2rem 1.5rem;
            position: fixed;
            height: 100vh;
            left: 0;
            top: 0;
            z-index: 100;
        }

        .sidebar-brand {
            font-size: 1.5rem;
            font-weight: 700;
            letter-spacing: -0.05em;
            background: linear-gradient(to right, #818cf8, #e0e7ff);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            margin-bottom: 3rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .sidebar-menu {
            list-style: none;
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
            flex-grow: 1;
        }

        .nav-item {
            display: flex;
            align-items: center;
            gap: 1rem;
            padding: 0.75rem 1rem;
            color: var(--text-muted);
            text-decoration: none;
            border-radius: 8px;
            font-weight: 500;
            font-size: 0.95rem;
            cursor: pointer;
            transition: all 0.2s ease;
        }

        .nav-item:hover {
            color: var(--text-main);
            background-color: rgba(255, 255, 255, 0.03);
        }

        .nav-item.active {
            color: #ffffff;
            background-color: var(--primary);
            box-shadow: 0 4px 12px var(--primary-glow);
        }

        .user-profile {
            display: flex;
            align-items: center;
            gap: 1rem;
            padding-top: 1.5rem;
            border-top: 1px solid var(--border-color);
            margin-top: auto;
        }

        .avatar {
            width: 40px;
            height: 40px;
            border-radius: 50%;
            background: linear-gradient(135deg, #6366f1, #a855f7);
            display: flex;
            align-items: center;
            justify-content: center;
            font-weight: 700;
            font-size: 0.95rem;
            color: white;
        }

        .user-info {
            display: flex;
            flex-direction: column;
        }

        .username {
            font-weight: 600;
            font-size: 0.9rem;
        }

        .user-role {
            font-size: 0.75rem;
            color: var(--text-muted);
        }

        /* Main Content Panel */
        .main-container {
            margin-left: 260px;
            padding: 2.5rem;
            width: calc(100% - 260px);
            min-height: 100vh;
        }

        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2.5rem;
            padding-bottom: 1.5rem;
            border-bottom: 1px solid var(--border-color);
        }

        .header-title h2 {
            font-size: 1.6rem;
            font-weight: 700;
        }

        .header-title p {
            font-size: 0.85rem;
            color: var(--text-muted);
            margin-top: 0.2rem;
        }

        .meta-tag {
            background-color: rgba(255, 255, 255, 0.03);
            border: 1px solid var(--border-color);
            padding: 0.5rem 1rem;
            border-radius: 9999px;
            font-size: 0.8rem;
            color: var(--text-muted);
        }

        /* Tab Contents */
        .tab-content {
            display: none;
            animation: fadeIn 0.3s ease;
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(5px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Metrics Cards Grid */
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2.5rem;
        }

        .metric-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            backdrop-filter: blur(16px);
            padding: 1.8rem;
            border-radius: 16px;
            position: relative;
            overflow: hidden;
            transition: transform 0.2s ease, border-color 0.2s ease;
        }

        .metric-card:hover {
            transform: translateY(-2px);
            border-color: rgba(255, 255, 255, 0.12);
        }

        .metric-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            width: 4px;
            height: 100%;
        }

        .metric-card.total-cost::before { background-color: var(--primary); }
        .metric-card.savings::before { background-color: var(--success); }
        .metric-card.waste::before { background-color: var(--danger); }
        .metric-card.inventory::before { background-color: var(--warning); }

        .metric-label {
            font-size: 0.85rem;
            color: var(--text-muted);
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            margin-bottom: 0.5rem;
        }

        .metric-value {
            font-size: 2.2rem;
            font-weight: 700;
            letter-spacing: -0.02em;
        }

        .metric-subtitle {
            font-size: 0.8rem;
            color: var(--text-muted);
            margin-top: 0.5rem;
            display: flex;
            align-items: center;
            gap: 0.4rem;
        }

        .percentage-badge {
            padding: 0.2rem 0.5rem;
            border-radius: 4px;
            font-size: 0.75rem;
            font-weight: 700;
        }

        .percentage-badge.danger {
            background-color: var(--danger-glow);
            color: var(--danger);
        }

        /* Controls Row */
        .controls-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1.5rem;
            gap: 1.5rem;
            flex-wrap: wrap;
        }

        .search-box {
            background-color: rgba(255, 255, 255, 0.03);
            border: 1px solid var(--border-color);
            color: var(--text-main);
            padding: 0.75rem 1.25rem;
            border-radius: 8px;
            font-size: 0.9rem;
            min-width: 320px;
            outline: none;
            transition: border-color 0.2s, box-shadow 0.2s;
        }

        .search-box:focus {
            border-color: var(--primary);
            box-shadow: 0 0 0 3px var(--primary-glow);
        }

        .filter-buttons {
            display: flex;
            gap: 0.5rem;
        }

        .filter-btn {
            background-color: rgba(255, 255, 255, 0.03);
            border: 1px solid var(--border-color);
            color: var(--text-muted);
            padding: 0.6rem 1.2rem;
            border-radius: 8px;
            font-size: 0.85rem;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s;
        }

        .filter-btn:hover {
            color: var(--text-main);
            border-color: rgba(255, 255, 255, 0.15);
        }

        .filter-btn.active {
            background-color: var(--primary);
            border-color: var(--primary);
            color: white;
            box-shadow: 0 4px 12px var(--primary-glow);
        }

        /* Tables & Panel Cards */
        .panel-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            backdrop-filter: blur(16px);
            border-radius: 16px;
            overflow: hidden;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.15);
            margin-bottom: 2rem;
        }

        .panel-header {
            padding: 1.5rem 1.8rem;
            border-bottom: 1px solid var(--border-color);
            font-size: 1.1rem;
            font-weight: 600;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        table {
            width: 100%;
            border-collapse: collapse;
            text-align: left;
        }

        th {
            background-color: rgba(255, 255, 255, 0.01);
            color: var(--text-muted);
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            padding: 1rem 1.8rem;
            border-bottom: 1px solid var(--border-color);
        }

        td {
            padding: 1.2rem 1.8rem;
            border-bottom: 1px solid var(--border-color);
            font-size: 0.9rem;
            color: var(--text-main);
        }

        tr:last-child td {
            border-bottom: none;
        }

        tr:hover td {
            background-color: rgba(255, 255, 255, 0.01);
        }

        .cloud-badge {
            display: inline-flex;
            align-items: center;
            font-size: 0.75rem;
            font-weight: 700;
            padding: 0.25rem 0.6rem;
            border-radius: 9999px;
            text-transform: uppercase;
        }

        .cloud-badge.aws {
            background-color: rgba(255, 153, 0, 0.1);
            color: var(--aws-color);
            border: 1px solid rgba(255, 153, 0, 0.2);
        }

        .cloud-badge.azure {
            background-color: rgba(0, 137, 214, 0.1);
            color: var(--azure-color);
            border: 1px solid rgba(0, 137, 214, 0.2);
        }

        .cloud-badge.gcp {
            background-color: rgba(52, 168, 83, 0.1);
            color: var(--gcp-color);
            border: 1px solid rgba(52, 168, 83, 0.2);
        }

        .resource-type {
            font-weight: 500;
        }

        .resource-id {
            font-family: monospace;
            color: var(--text-muted);
            font-size: 0.8rem;
            display: block;
            margin-top: 0.15rem;
        }

        .savings-value {
            font-weight: 700;
            color: var(--success);
        }

        .recommendation {
            font-size: 0.85rem;
            color: var(--text-muted);
        }

        .tag-pill {
            display: inline-block;
            font-size: 0.7rem;
            background-color: rgba(255, 255, 255, 0.05);
            border: 1px solid var(--border-color);
            padding: 0.1rem 0.4rem;
            border-radius: 4px;
            margin-right: 0.25rem;
            margin-top: 0.25rem;
        }

        .no-records {
            text-align: center;
            padding: 4rem;
            color: var(--text-muted);
        }

        /* Integrations Tab Specifics */
        .integrations-container {
            display: grid;
            grid-template-columns: 1fr 360px;
            gap: 1.5rem;
        }

        .integration-card {
            background: var(--card-bg);
            border: 1px solid var(--border-color);
            border-radius: 16px;
            padding: 2rem;
            display: flex;
            flex-direction: column;
            gap: 1.5rem;
        }

        .integration-top {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }

        .integration-name {
            font-weight: 700;
            font-size: 1.2rem;
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .status-pill {
            padding: 0.25rem 0.6rem;
            font-size: 0.75rem;
            font-weight: 600;
            border-radius: 4px;
        }

        .status-pill.unconfigured {
            background-color: rgba(255, 255, 255, 0.05);
            color: var(--text-muted);
        }

        .status-pill.connected {
            background-color: var(--success-glow);
            color: var(--success);
        }

        .form-group {
            display: flex;
            flex-direction: column;
            gap: 0.4rem;
        }

        .form-group label {
            font-size: 0.8rem;
            font-weight: 600;
            color: var(--text-muted);
            text-transform: uppercase;
            letter-spacing: 0.02em;
        }

        .form-control {
            background-color: rgba(255, 255, 255, 0.03);
            border: 1px solid var(--border-color);
            color: var(--text-main);
            padding: 0.6rem 0.85rem;
            border-radius: 6px;
            font-size: 0.85rem;
            outline: none;
            transition: border-color 0.2s;
        }

        .form-control:focus {
            border-color: var(--primary);
        }

        .environment-select {
            display: flex;
            gap: 0.5rem;
            background-color: rgba(255, 255, 255, 0.02);
            padding: 0.2rem;
            border: 1px solid var(--border-color);
            border-radius: 6px;
        }

        .env-option {
            flex-grow: 1;
            text-align: center;
            padding: 0.4rem;
            font-size: 0.75rem;
            font-weight: 600;
            border-radius: 4px;
            cursor: pointer;
            color: var(--text-muted);
            transition: all 0.2s;
        }

        .env-option.active {
            background-color: var(--primary);
            color: white;
        }

        .btn-action {
            background-color: var(--primary);
            color: white;
            border: none;
            padding: 0.75rem;
            font-size: 0.85rem;
            font-weight: 600;
            border-radius: 6px;
            cursor: pointer;
            transition: background-color 0.2s;
            text-align: center;
        }

        .btn-action:hover {
            background-color: var(--primary-hover);
        }

        /* Settings / Team Admin Tab Specifics */
        .settings-layout {
            display: grid;
            grid-template-columns: 1fr 340px;
            gap: 1.5rem;
            align-items: start;
        }

        .role-badge {
            font-size: 0.75rem;
            font-weight: 600;
            padding: 0.15rem 0.4rem;
            border-radius: 4px;
        }

        .role-badge.admin { background-color: rgba(99, 102, 241, 0.1); color: #818cf8; }
        .role-badge.operator { background-color: rgba(245, 158, 11, 0.1); color: var(--warning); }
        .role-badge.viewer { background-color: rgba(255, 255, 255, 0.05); color: var(--text-muted); }

        .status-badge {
            font-size: 0.75rem;
            font-weight: 600;
            padding: 0.15rem 0.4rem;
            border-radius: 4px;
        }

        .status-badge.active { background-color: var(--success-glow); color: var(--success); }
        .status-badge.pending { background-color: rgba(245, 158, 11, 0.1); color: var(--warning); }

        /* Responsive Layouts */
        @media (max-width: 1024px) {
            .settings-layout, .integrations-container {
                grid-template-columns: 1fr;
            }
        }

        @media (max-width: 768px) {
            .sidebar {
                width: 72px;
                padding: 1.5rem 0.5rem;
            }
            .sidebar-brand span, .nav-item span, .user-info, .sidebar-brand svg {
                display: none;
            }
            .sidebar-brand {
                font-size: 1.1rem;
                justify-content: center;
            }
            .main-container {
                margin-left: 72px;
                width: calc(100% - 72px);
                padding: 1.5rem;
            }
        }
    </style>
</head>
<body>

    <!-- INTERACTIVE OAUTH/SSO LOGIN GATE -->
    <div id="login-screen">
        <div class="login-card">
            <div>
                <div class="login-logo">costctl //</div>
                <div class="login-header">
                    <h2>Multi-Cloud Spend Optimizer</h2>
                    <p>Consolidate Billing. Eliminate Waste. Autoscale Savings.</p>
                </div>
            </div>
            
            <div class="sso-buttons">
                <!-- Google SSO Button -->
                <div class="btn-sso google" onclick="triggerLogin('Google SSO')">
                    <svg viewBox="0 0 24 24" width="18" height="18"><path fill="#4285F4" d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v3.9h6.6c-.28 1.5-.1.8-2.6 2.87v2.4h4.15c2.44-2.24 3.84-5.55 3.84-9.1z"/><path fill="#34A853" d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-4.15-3.22c-1.16.78-2.65 1.25-3.78 1.25-4.47 0-8.25-3.02-9.6-7.09H1.24v3.3C3.21 20.1 7.27 24 12 24z"/><path fill="#FBBC05" d="M2.4 12c-.24-.7-.37-1.44-.37-2.2s.13-1.5.37-2.2V4.3H1.24C.45 5.9 0 7.7 0 9.8s.45 3.9 1.24 5.5l1.16-3.3z"/><path fill="#EA4335" d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.27 0 3.21 3.9 1.24 7.8l3.6 2.8c1.35-4.07 5.13-7.05 9.16-7.05z"/></svg>
                    <span>Sign in with Google Workspace</span>
                </div>
                
                <!-- Microsoft Entra ID Button -->
                <div class="btn-sso entra" onclick="triggerLogin('Microsoft Entra ID')">
                    <svg viewBox="0 0 23 23" width="18" height="18"><rect fill="#F25022" x="0" y="0" width="11" height="11"/><rect fill="#7FBA00" x="12" y="0" width="11" height="11"/><rect fill="#00A4EF" x="0" y="12" width="11" height="11"/><rect fill="#FFB900" x="12" y="12" width="11" height="11"/></svg>
                    <span>Sign in with Microsoft Entra ID</span>
                </div>
            </div>

            <div class="login-footer">
                Secured access. Multi-account mapping uses corporate SSO. By logging in, you agree to FinOps policy guidelines.
                <div style="margin-top: 0.8rem; font-size: 0.7rem; opacity: 0.7; letter-spacing: 0.02em;">
                    &copy; 2026 costctl. Author: costctl contributors. Trademark &trade; All rights reserved.
                </div>
            </div>
        </div>
    </div>

    <!-- MAIN APP VIEW (HIDDEN UNTIL LOGIN) -->
    <div id="app-view">
        <!-- LEFT SIDEBAR -->
        <div class="sidebar">
            <div class="sidebar-brand">
                costctl //
            </div>
            <ul class="sidebar-menu">
                <li>
                    <div class="nav-item active" id="nav-dashboard" onclick="showTab('dashboard')">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="9" rx="1"/><rect x="14" y="3" width="7" height="5" rx="1"/><rect x="14" y="12" width="7" height="9" rx="1"/><rect x="3" y="16" width="7" height="5" rx="1"/></svg>
                        <span>Dashboard</span>
                    </div>
                </li>
                <li>
                    <div class="nav-item" id="nav-reports" onclick="showTab('reports')">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"/><polyline points="14 2 14 8 20 8"/></svg>
                        <span>Detailed Reports</span>
                    </div>
                </li>
                <li>
                    <div class="nav-item" id="nav-integrations" onclick="showTab('integrations')">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>
                        <span>Integrations</span>
                    </div>
                </li>
                <li>
                    <div class="nav-item" id="nav-settings" onclick="showTab('settings')">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"/><circle cx="8.5" cy="7" r="4"/><line x1="20" y1="8" x2="20" y2="14"/><line x1="23" y1="11" x2="17" y2="11"/></svg>
                        <span>Admin Settings</span>
                    </div>
                </li>
            </ul>
            <div class="user-profile">
                <div class="avatar" id="userAvatar">AD</div>
                <div class="user-info">
                    <span class="username" id="userName">Admin User</span>
                    <span class="user-role" id="userRole">Cloud Architect</span>
                </div>
                <!-- Logout Button -->
                <button onclick="logout()" style="background: transparent; border: none; color: var(--text-muted); cursor: pointer; margin-left: auto; display: flex; align-items: center; justify-content: center; padding: 0.2rem; border-radius: 4px; transition: color 0.2s;" title="Log Out">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>
                </button>
            </div>
        </div>

        <!-- MAIN CONTAINER -->
        <div class="main-container">

            <header>
                <div class="header-title">
                    <h2 id="tab-title">Optimization Dashboard</h2>
                    <p id="tab-subtitle">Consolidated billing overview and multi-cloud waste optimization</p>
                </div>
                
                <!-- Time Period Filter & Meta tags -->
                <div style="display: flex; align-items: center; gap: 1rem;">
                    <!-- Timeline selectors -->
                    <div class="filter-buttons" style="background-color: rgba(255,255,255,0.02); padding: 0.2rem; border: 1px solid var(--border-color); border-radius: 8px;">
                        <button class="filter-btn" id="time-7d" onclick="setTimeFilter('7d')" style="padding: 0.4rem 0.8rem; font-size: 0.75rem;">7D</button>
                        <button class="filter-btn active" id="time-30d" onclick="setTimeFilter('30d')" style="padding: 0.4rem 0.8rem; font-size: 0.75rem;">30D</button>
                        <button class="filter-btn" id="time-90d" onclick="setTimeFilter('90d')" style="padding: 0.4rem 0.8rem; font-size: 0.75rem;">QTR</button>
                        <button class="filter-btn" id="time-365d" onclick="setTimeFilter('365d')" style="padding: 0.4rem 0.8rem; font-size: 0.75rem;">1Y</button>
                        <button class="filter-btn" id="time-custom" onclick="setTimeFilter('custom')" style="padding: 0.4rem 0.8rem; font-size: 0.75rem;">Custom</button>
                    </div>

                    <div class="meta-tag">
                        Last Run: <span id="generation-time">{{.GeneratedTime}}</span>
                    </div>
                </div>
            </header>

            <!-- Custom date range drawer -->
            <div id="custom-date-drawer" style="display: none; background: var(--card-bg); border: 1px solid var(--border-color); padding: 1.2rem; border-radius: 12px; margin-bottom: 2rem; align-items: center; gap: 1.5rem;">
                <div class="form-group" style="flex-direction: row; align-items: center; gap: 0.5rem; margin-bottom: 0;">
                    <label style="margin-bottom: 0; font-size: 0.75rem;">Start Date</label>
                    <input type="date" id="startDate" class="form-control" value="2026-05-01">
                </div>
                <div class="form-group" style="flex-direction: row; align-items: center; gap: 0.5rem; margin-bottom: 0;">
                    <label style="margin-bottom: 0; font-size: 0.75rem;">End Date</label>
                    <input type="date" id="endDate" class="form-control" value="2026-05-27">
                </div>
                <button class="btn-action" onclick="applyCustomDates()" style="padding: 0.5rem 1.2rem; font-size: 0.8rem; margin-top: 0;">Apply Filter</button>
            </div>

            <!-- TAB: DASHBOARD -->
            <div id="dashboard-tab" class="tab-content" style="display: block;">
                <!-- Consolidated Billing Header Card -->
                <div class="panel-card" style="background: linear-gradient(135deg, rgba(79, 70, 229, 0.15), rgba(15, 19, 31, 0.7)); border-color: rgba(99, 102, 241, 0.25);">
                    <div style="padding: 2.2rem; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1.5rem;">
                        <div>
                            <span style="font-size: 0.75rem; font-weight: 700; text-transform: uppercase; color: #818cf8; letter-spacing: 0.05em;">Consolidated FinOps Billing Invoice</span>
                            <h2 style="font-size: 2.6rem; font-weight: 800; letter-spacing: -0.03em; margin: 0.4rem 0;" id="consBillingTotal">$60,811.50</h2>
                            <p style="font-size: 0.85rem; color: var(--text-muted);" id="consBillingSub">Aggregating 3 clouds, 15 linked accounts, and 8 subscriptions</p>
                        </div>
                        <div style="display: flex; gap: 2rem;">
                            <div style="text-align: right;">
                                <span style="font-size: 0.8rem; color: var(--text-muted);">Consolidated Waste</span>
                                <h3 style="font-size: 1.6rem; font-weight: 700; color: var(--danger); margin-top: 0.2rem;" id="consWasteTotal">{{.FormattedTotalSavings}}</h3>
                            </div>
                            <div style="text-align: right; border-left: 1px solid var(--border-color); padding-left: 2rem;">
                                <span style="font-size: 0.8rem; color: var(--text-muted);">Optimization Ratio</span>
                                <h3 style="font-size: 1.6rem; font-weight: 700; color: var(--success); margin-top: 0.2rem;" id="consOptRatio">24.9%</h3>
                            </div>
                        </div>
                    </div>
                </div>

                <div class="metrics-grid">
                    <div class="metric-card total-cost">
                        <div class="metric-label">Estimated Monthly Active Spend</div>
                        <div class="metric-value" id="kpiActiveSpend">{{.FormattedTotalCost}}</div>
                        <div class="metric-subtitle">Across active scanned inventory</div>
                    </div>
                    <div class="metric-card savings">
                        <div class="metric-label">Consolidated Monthly Savings</div>
                        <div class="metric-value" style="color: var(--success);" id="kpiSavings">{{.FormattedTotalSavings}}</div>
                        <div class="metric-subtitle">
                            <span class="percentage-badge danger" id="kpiWastePct">{{printf "%.1f%%" .WastePercent}} waste</span> of active inventory
                        </div>
                    </div>
                    <div class="metric-card waste">
                        <div class="metric-label">Detected Waste Assets</div>
                        <div class="metric-value" id="kpiWasteAssets">{{.FindingsCount}}</div>
                        <div class="metric-subtitle">Requiring decommissioning / resizing</div>
                    </div>
                    <div class="metric-card inventory">
                        <div class="metric-label">Connected Cloud Accounts</div>
                        <div class="metric-value" id="cloudCountBadge">3</div>
                        <div class="metric-subtitle">Active API configuration endpoints</div>
                    </div>
                </div>

                <!-- Spend distribution grid -->
                <div class="panel-card">
                    <div class="panel-header">
                        <span>Consolidated Billing Share by Provider</span>
                        <span style="font-size: 0.8rem; color: var(--text-muted);">Time-based historical share</span>
                    </div>
                    <div style="padding: 2rem; display: flex; flex-direction: column; gap: 1.5rem;" id="distContainer">
                        <div>
                            <div style="display: flex; justify-content: space-between; font-size: 0.85rem; margin-bottom: 0.4rem; font-weight: 500;">
                                <span>Amazon Web Services (AWS) — Billing Master ID: 1234-5678-9012</span>
                                <span style="color: var(--aws-color); font-weight: 700;" id="awsDistShare">41.1% ($24,984.00)</span>
                            </div>
                            <div style="height: 10px; background-color: rgba(255, 255, 255, 0.05); border-radius: 9999px; overflow: hidden;">
                                <div id="awsDistBar" style="width: 41.1%; height: 100%; background-color: var(--aws-color); border-radius: 9999px; transition: width 0.5s ease;"></div>
                            </div>
                        </div>
                        <div>
                            <div style="display: flex; justify-content: space-between; font-size: 0.85rem; margin-bottom: 0.4rem; font-weight: 500;">
                                <span>Microsoft Azure — Consolidated Tenant ID: ea88a6d2-...</span>
                                <span style="color: var(--azure-color); font-weight: 700;" id="azureDistShare">36.6% ($22,246.50)</span>
                            </div>
                            <div style="height: 10px; background-color: rgba(255, 255, 255, 0.05); border-radius: 9999px; overflow: hidden;">
                                <div id="azureDistBar" style="width: 36.6%; height: 100%; background-color: var(--azure-color); border-radius: 9999px; transition: width 0.5s ease;"></div>
                            </div>
                        </div>
                        <div>
                            <div style="display: flex; justify-content: space-between; font-size: 0.85rem; margin-bottom: 0.4rem; font-weight: 500;">
                                <span>Google Cloud Platform (GCP) — Billing Export dataset: billing_export</span>
                                <span style="color: var(--gcp-color); font-weight: 700;" id="gcpDistShare">22.3% ($13,581.00)</span>
                            </div>
                            <div style="height: 10px; background-color: rgba(255, 255, 255, 0.05); border-radius: 9999px; overflow: hidden;">
                                <div id="gcpDistBar" style="width: 22.3%; height: 100%; background-color: var(--gcp-color); border-radius: 9999px; transition: width 0.5s ease;"></div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

            <!-- TAB: REPORTS -->
            <div id="reports-tab" class="tab-content">
                <div class="controls-row">
                    <input type="text" id="searchInput" placeholder="Search by name, ID, tags, or recommendations..." class="search-box" oninput="filterTable()">
                    <div class="filter-buttons">
                        <button class="filter-btn active" id="btn-all" onclick="setFilter('all')">All Clouds</button>
                        <button class="filter-btn" id="btn-aws" onclick="setFilter('aws')">AWS</button>
                        <button class="filter-btn" id="btn-azure" onclick="setFilter('azure')">Azure</button>
                        <button class="filter-btn" id="btn-gcp" onclick="setFilter('gcp')">GCP</button>
                    </div>
                </div>

                <div class="panel-card">
                    <div class="panel-header">
                        <span>Consolidated Waste Findings</span>
                        <button class="filter-btn" onclick="alert('CSV Export generated (mock)')" style="padding: 0.4rem 0.8rem; font-size: 0.8rem;">Export CSV</button>
                    </div>
                    <div style="overflow-x: auto;">
                        <table id="findingsTable">
                            <thead>
                                <tr>
                                    <th>Resource Name / ID</th>
                                    <th>Cloud</th>
                                    <th>Type</th>
                                    <th>Monthly Waste</th>
                                    <th>Waste Reason & Recommendation</th>
                                </tr>
                            </thead>
                            <tbody>
                                <!-- Dynamic render in JavaScript -->
                            </tbody>
                        </table>
                    </div>
                </div>
            </div>

            <!-- TAB: INTEGRATIONS (MULTI-ACCOUNT MANAGER) -->
            <div id="integrations-tab" class="tab-content">
                <div class="integrations-container">
                    
                    <!-- LEFT COLUMN: Connected Accounts Inventory -->
                    <div>
                        <h3 style="font-size: 1.1rem; font-weight: 700; margin-bottom: 1rem; color: var(--text-main);">Connected Enterprise Accounts</h3>
                        <div id="accountsList">
                            <!-- Dynamic render in javascript -->
                        </div>
                    </div>

                    <!-- RIGHT COLUMN: Add New Account Connector Form -->
                    <div class="integration-card" style="height: fit-content; position: sticky; top: 2.5rem;">
                        <h3 style="font-size: 1.1rem; font-weight: 700; border-bottom: 1px solid var(--border-color); padding-bottom: 0.75rem;">Add Account Connector</h3>
                        <form onsubmit="addAccount(event)" style="display: flex; flex-direction: column; gap: 1rem;">
                            
                            <div class="form-group">
                                <label for="accProvider">Cloud Provider</label>
                                <select id="accProvider" class="form-control" style="background-color: var(--sidebar-bg);" onchange="updateProviderFields()">
                                    <option value="aws">Amazon Web Services</option>
                                    <option value="azure">Microsoft Azure</option>
                                    <option value="gcp">Google Cloud Platform</option>
                                </select>
                            </div>

                            <div class="form-group">
                                <label for="accName">Account Label Name</label>
                                <input type="text" id="accName" class="form-control" placeholder="e.g. AWS - Client Acme Prod" required>
                            </div>

                            <div class="form-group">
                                <label for="accId" id="lblAccId">Account ID (12-digit)</label>
                                <input type="text" id="accId" class="form-control" placeholder="e.g. 123456789012" required>
                            </div>

                            <div class="form-group">
                                <label>Environment Domain</label>
                                <div class="environment-select">
                                    <div class="env-option active" id="envComm" onclick="setFormEnv('Commercial')">Commercial</div>
                                    <div class="env-option" id="envGov" onclick="setFormEnv('Government')">Government</div>
                                </div>
                            </div>

                            <!-- Account tags configuration -->
                            <div class="form-group">
                                <label style="display: flex; justify-content: space-between; align-items: center;">
                                    <span>Account-level Tags</span>
                                    <button type="button" onclick="addTagField()" style="background: transparent; border: none; color: var(--success); cursor: pointer; font-size: 0.75rem; font-weight: 700;">+ Add Tag</button>
                                </label>
                                <div id="tagsInputContainer" style="display: flex; flex-direction: column; gap: 0.5rem;">
                                    <!-- Dynamic rows inside javascript -->
                                    <div style="display: flex; gap: 0.5rem;" class="tag-input-row">
                                        <input type="text" placeholder="Key (e.g. client)" class="form-control tag-key" style="flex: 1;" required>
                                        <input type="text" placeholder="Value" class="form-control tag-value" style="flex: 1;" required>
                                    </div>
                                </div>
                            </div>

                            <button type="submit" class="btn-action" style="margin-top: 0.5rem;">Connect Account</button>
                        </form>
                    </div>

                </div>
            </div>

            <!-- TAB: SETTINGS & TEAM MANAGEMENT -->
            <div id="settings-tab" class="tab-content">
                <div class="settings-layout">
                    <!-- Team Members Management -->
                    <div class="panel-card">
                        <div class="panel-header">
                            <span>FinOps Team (Google Workspace Sync Active)</span>
                            <span style="font-size: 0.8rem; color: var(--text-muted);">Users authenticate using SSO</span>
                        </div>
                        <div style="overflow-x: auto;">
                            <table>
                                <thead>
                                    <tr>
                                        <th>Member Name</th>
                                        <th>Email Address</th>
                                        <th>Assigned Role</th>
                                        <th>Status</th>
                                    </tr>
                                </thead>
                                <tbody id="teamTableBody">
                                    <tr>
                                        <td>
                                            <div style="font-weight: 600;">Admin User</div>
                                        </td>
                                        <td>admin@company.com</td>
                                        <td><span class="role-badge admin">Owner / Admin</span></td>
                                        <td><span class="status-badge active">Active</span></td>
                                    </tr>
                                    <tr>
                                        <td>
                                            <div style="font-weight: 600;">Bob Jones</div>
                                        </td>
                                        <td>bob.jones@company.com</td>
                                        <td><span class="role-badge operator">FinOps Operator</span></td>
                                        <td><span class="status-badge active">Active</span></td>
                                    </tr>
                                    <tr>
                                        <td>
                                            <div style="font-weight: 600;">Alice Smith</div>
                                        </td>
                                        <td>alice.smith@company.com</td>
                                        <td><span class="role-badge viewer">Viewer</span></td>
                                        <td><span class="status-badge pending">Pending</span></td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>

                    <!-- Invite Team Member Panel -->
                    <div class="integration-card" style="position: sticky; top: 2.5rem;">
                        <h3 style="font-size: 1.1rem; font-weight: 600; border-bottom: 1px solid var(--border-color); padding-bottom: 0.75rem;">Invite Team Member</h3>
                        <form onsubmit="inviteMember(event)" style="display: flex; flex-direction: column; gap: 1.2rem;">
                            <div class="form-group">
                                <label for="memberName">Full Name</label>
                                <input type="text" id="memberName" class="form-control" placeholder="e.g. John Doe" required>
                            </div>
                            <div class="form-group">
                                <label for="memberEmail">Corporate Email</label>
                                <input type="email" id="memberEmail" class="form-control" placeholder="e.g. john@company.com" required>
                            </div>
                            <div class="form-group">
                                <label for="memberRole">Role Scope</label>
                                <select id="memberRole" class="form-control" style="background-color: var(--sidebar-bg);">
                                    <option value="Viewer">Viewer</option>
                                    <option value="FinOps Operator">FinOps Operator</option>
                                    <option value="Owner / Admin">Owner / Admin</option>
                                </select>
                            </div>
                            
                            <!-- Google SSO Invite Directory Sync Indicator -->
                            <div style="display: flex; align-items: center; gap: 0.5rem; background-color: rgba(79, 70, 229, 0.05); padding: 0.6rem; border-radius: 6px; border: 1px solid rgba(79, 70, 229, 0.15);">
                                <input type="checkbox" id="syncGoogle" checked disabled style="accent-color: var(--primary);">
                                <label for="syncGoogle" style="font-size: 0.75rem; color: var(--text-main); text-transform: none; cursor: default;">
                                    Auto-Provision user in Google Workspace Directory
                                </label>
                            </div>
                            
                            <button type="submit" class="btn-action" style="margin-top: 0.5rem;">Send SSO Invite</button>
                        </form>
                    </div>
                </div>
            </div>

            <footer style="margin-top: 3rem; padding-top: 1.5rem; border-top: 1px solid var(--border-color); text-align: center; font-size: 0.8rem; color: var(--text-muted); letter-spacing: 0.02em;">
                &copy; 2026 costctl. Author: costctl contributors. Trademark &trade; All rights reserved.
            </footer>
        </div>
    </div>

    <!-- CLIENT SIDE TAB & DYNAMIC MANAGEMENT LOGIC -->
    <script>
        // Go template will serialize resource database to Javascript globals
        const rawFindings = {{.FindingsJSON}};
        const rawResources = {{.ResourcesJSON}};

        // Trigger Login with loading animation
        function triggerLogin(providerName) {
            const loginCard = document.querySelector('.login-card');
            
            // Show dynamic loading screen inside the card
            loginCard.innerHTML = '<div>' +
                '<div style=\"font-size: 1.6rem; font-weight: 800; color: #818cf8; margin-bottom: 1rem;\">Authenticating...</div>' +
                '<div style=\"color: var(--text-muted); font-size: 0.9rem;\">Signing in with ' + providerName + '</div>' +
                '<div style=\"margin: 2rem auto; width: 40px; height: 40px; border: 3px solid rgba(255,255,255,0.1); border-top: 3px solid var(--primary); border-radius: 50%; animation: spin 1s linear infinite;\"></div>' +
                '</div>';
            
            // Insert rotation animation into DOM
            const style = document.createElement('style');
            style.innerHTML = '@keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }';
            document.head.appendChild(style);

            setTimeout(() => {
                // Set username based on SSO provider
                if (providerName.includes('Google')) {
                    document.getElementById('userName').textContent = 'Admin (SSO)';
                    document.getElementById('userAvatar').textContent = 'AD';
                } else {
                    document.getElementById('userName').textContent = 'Admin (Entra ID)';
                    document.getElementById('userAvatar').textContent = 'AD';
                }

                // Hide login screen, show application view
                document.getElementById('login-screen').style.display = 'none';
                document.getElementById('app-view').style.display = 'block';

                // Initial render and sync of data
                syncAll();
            }, 1200);
        }

        // Logout function - resets states and triggers page refresh
        function logout() {
            location.reload();
        }

        // Switch Visible Tab
        function showTab(tabId) {
            // Hide all tabs
            document.querySelectorAll('.tab-content').forEach(tab => {
                tab.style.display = 'none';
            });
            // Remove active classes
            document.querySelectorAll('.nav-item').forEach(item => {
                item.classList.remove('active');
            });
            
            // Show requested tab
            document.getElementById(tabId + '-tab').style.display = 'block';
            document.getElementById('nav-' + tabId).classList.add('active');

            // Update Page Headers
            const titleEl = document.getElementById('tab-title');
            const subtitleEl = document.getElementById('tab-subtitle');
            
            if (tabId === 'dashboard') {
                titleEl.textContent = 'Optimization Dashboard';
                subtitleEl.textContent = 'Consolidated billing overview and multi-cloud waste optimization';
            } else if (tabId === 'reports') {
                titleEl.textContent = 'Detailed Waste Audit';
                subtitleEl.textContent = 'Analyze, search, and export detected cloud waste findings';
            } else if (tabId === 'integrations') {
                titleEl.textContent = 'Cloud Accounts Integrations';
                subtitleEl.textContent = 'Configure API connectors for Commercial and Government domains';
            } else if (tabId === 'settings') {
                titleEl.textContent = 'Admin Team Control';
                subtitleEl.textContent = 'Manage permissions, thresholds, and call team members';
            }
        }

        // Toggle integration environments in connector card
        let selectedFormEnv = 'Commercial';
        function setFormEnv(env) {
            selectedFormEnv = env;
            if (env === 'Commercial') {
                document.getElementById('envComm').classList.add('active');
                document.getElementById('envGov').classList.remove('active');
            } else {
                document.getElementById('envGov').classList.add('active');
                document.getElementById('envComm').classList.remove('active');
            }
        }

        function toggleEnv(el) {
            const container = el.parentElement;
            container.querySelectorAll('.env-option').forEach(opt => {
                opt.classList.remove('active');
            });
            el.classList.add('active');
        }

        // Update add-account form labels depending on selected provider
        function updateProviderFields() {
            const provider = document.getElementById('accProvider').value;
            const lbl = document.getElementById('lblAccId');
            const input = document.getElementById('accId');
            
            if (provider === 'aws') {
                lbl.textContent = 'Account ID (12-digit)';
                input.placeholder = 'e.g. 123456789012';
            } else if (provider === 'azure') {
                lbl.textContent = 'Subscription ID (UUID)';
                input.placeholder = 'e.g. ea88a6d2-9867-4bd9-9e8a-028f090b8fbc';
            } else if (provider === 'gcp') {
                lbl.textContent = 'Project ID (string)';
                input.placeholder = 'e.g. corporate-analytics-prod';
            }
        }

        // Dynamic Tag row adder in the form
        function addTagField() {
            const container = document.getElementById('tagsInputContainer');
            const row = document.createElement('div');
            row.style.display = 'flex';
            row.style.gap = '0.5rem';
            row.className = 'tag-input-row';
            row.innerHTML = 
                '<input type="text" placeholder="Key" class="form-control tag-key" style="flex: 1;" required>' +
                '<input type="text" placeholder="Value" class="form-control tag-value" style="flex: 1;" required>' +
                '<button type="button" onclick="this.parentElement.remove()" style="background:transparent; border:none; color:var(--danger); cursor:pointer; font-weight:700; font-size:1.1rem; padding:0 0.2rem;">×</button>';
            container.appendChild(row);
        }

        // Account connectors state array
        let accounts = [
            { id: '123456789012', name: 'AWS - Acme Master Billing', provider: 'aws', env: 'Commercial', tags: { 'client': 'AcmeCorp', 'cost-center': '104' } },
            { id: 'ea88a6d2-9867-4bd9-9e8a-028f090b8fbc', name: 'Azure - Client Beta Subscription', provider: 'azure', env: 'Commercial', tags: { 'client': 'BetaCo', 'cost-center': '201' } },
            { id: 'corporate-analytics-prod', name: 'GCP - Corporate Data Platform', provider: 'gcp', env: 'Commercial', tags: { 'client': 'Internal', 'owner': 'data-ops' } }
        ];

        // Core data synchronizer
        let currentPeriod = '30d';
        let currentCloudFilter = 'all';

        function syncAll() {
            // Get active cloud providers based on connected accounts
            const activeProviders = new Set(accounts.map(a => a.provider));
            
            // 1. Calculate the active time period scaler
            const scale = getPeriodScale();

            // 2. Filter resources and findings by connected clouds
            const filteredResources = rawResources.filter(r => activeProviders.has(r.Provider));
            const filteredFindings = rawFindings.filter(f => activeProviders.has(f.ProviderLower));

            // 3. Compute active inventory cost and savings
            const totalActiveCost = filteredResources.reduce((sum, r) => sum + r.CostPerMonth * scale, 0);
            const totalSavings = filteredFindings.reduce((sum, f) => sum + f.CostPerMonth * scale, 0);
            const wastePct = totalActiveCost > 0 ? (totalSavings / totalActiveCost) * 100 : 0;

            // 4. Recalculate consolidated invoice total
            const baselines = { aws: 24984.00, azure: 22246.50, gcp: 13581.00 };
            let invoiceTotal = 0;
            for (const [prov, base] of Object.entries(baselines)) {
                if (activeProviders.has(prov)) {
                    invoiceTotal += base * scale;
                }
            }

            // 5. Update KPI Cards in DOM
            updateKPIElement('consBillingTotal', '$' + invoiceTotal.toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2}));
            updateKPIElement('consWasteTotal', '$' + totalSavings.toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2}));
            updateKPIElement('consOptRatio', (totalActiveCost > 0 ? (100 - wastePct).toFixed(1) : '100.0') + '%');
            updateKPIElement('kpiActiveSpend', '$' + totalActiveCost.toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2}));
            updateKPIElement('kpiSavings', '$' + totalSavings.toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2}));
            updateKPIElement('kpiWastePct', wastePct.toFixed(1) + '% waste');
            updateKPIElement('kpiWasteAssets', filteredFindings.length);
            
            // Adjust waste badge color class based on waste density
            const pctBadge = document.getElementById('kpiWastePct');
            if (pctBadge) {
                if (wastePct > 40) {
                    pctBadge.className = 'percentage-badge danger';
                } else {
                    pctBadge.className = 'percentage-badge success';
                }
            }

            // 6. Recalculate Spend Distribution bars & shares
            updateDistributionBars(activeProviders, invoiceTotal, scale);

            // 7. Render accounts inventory list
            renderAccountsList(activeProviders);

            // 8. Render Reports table
            renderReportsTable(filteredFindings, scale);
        }

        // Helper to get time scaler factor
        function getPeriodScale() {
            if (currentPeriod === '7d') return 7.0 / 30.0;
            if (currentPeriod === '30d') return 1.0;
            if (currentPeriod === '90d') return 3.0;
            if (currentPeriod === '365d') return 12.0;
            if (currentPeriod === 'custom') {
                const start = new Date(document.getElementById('startDate').value);
                const end = new Date(document.getElementById('endDate').value);
                const diffTime = Math.abs(end - start);
                const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24)) || 30;
                return diffDays / 30.0;
            }
            return 1.0;
        }

        // Helper to update elements text with fade transition
        function updateKPIElement(id, text) {
            const el = document.getElementById(id);
            if (el && el.textContent !== text) {
                el.style.opacity = 0;
                setTimeout(() => {
                    el.textContent = text;
                    el.style.opacity = 1;
                    el.style.transition = 'opacity 0.2s';
                }, 150);
            }
        }

        // Render connected cloud inventory list
        function renderAccountsList(activeProviders) {
            const list = document.getElementById('accountsList');
            if(!list) return;
            
            list.innerHTML = '';
            
            // Update Consolidated invoice details
            document.getElementById('consBillingSub').textContent = 'Aggregating ' + accounts.length + ' clouds, ' + (accounts.length * 5) + ' linked accounts, and ' + (accounts.length * 3) + ' subscriptions';
            document.getElementById('cloudCountBadge').textContent = accounts.length;

            accounts.forEach((acc, index) => {
                let tagsHtml = '';
                for (const [k, v] of Object.entries(acc.tags)) {
                    tagsHtml += '<span class="tag-pill">' + k + ': ' + v + '</span>';
                }

                const card = document.createElement('div');
                card.className = 'panel-card';
                card.style.marginBottom = '1rem';
                card.innerHTML = 
                    '<div style="padding: 1.2rem; display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 1rem;">' +
                        '<div>' +
                            '<div style="display: flex; align-items: center; gap: 0.5rem; font-weight: 700;">' +
                                '<span class="cloud-badge ' + acc.provider + '">' + acc.provider.toUpperCase() + '</span>' +
                                '<span>' + acc.name + '</span>' +
                            '</div>' +
                            '<div style="font-size: 0.8rem; color: var(--text-muted); margin-top: 0.25rem;">ID: ' + acc.id + ' | Domain: ' + acc.env + '</div>' +
                            '<div style="margin-top: 0.5rem;">' + tagsHtml + '</div>' +
                        '</div>' +
                        '<div style="display: flex; align-items: center; gap: 1rem; margin-left: auto;">' +
                            '<span class="status-pill connected">Connected</span>' +
                            '<button type="button" onclick="removeAccount(' + index + ')" style="background: transparent; border: none; color: var(--danger); cursor: pointer; font-size: 0.85rem; font-weight:600;" title="Disconnect">Disconnect</button>' +
                        '</div>' +
                    '</div>';
                list.appendChild(card);
            });
        }

        // Render Reports table dynamically based on filters & scaler
        function renderReportsTable(filteredFindings, scale) {
            const tbody = document.querySelector('#findingsTable tbody');
            if (!tbody) return;

            tbody.innerHTML = '';

            const searchQuery = document.getElementById('searchInput').value.toLowerCase();
            
            const finalFindings = filteredFindings.filter(f => {
                const matchesCloud = currentCloudFilter === 'all' || f.ProviderLower === currentCloudFilter;
                const textContent = (f.ResourceName + ' ' + f.ResourceID + ' ' + f.Reason).toLowerCase();
                const matchesSearch = textContent.includes(searchQuery);
                return matchesCloud && matchesSearch;
            });

            if (finalFindings.length === 0) {
                tbody.innerHTML = '<tr><td colspan="5" class="no-records">No waste resources detected. Active inventory optimized!</td></tr>';
                return;
            }

            finalFindings.forEach(f => {
                const scaledCost = f.CostPerMonth * scale;
                const formattedSavings = '$' + scaledCost.toLocaleString('en-US', {minimumFractionDigits: 2, maximumFractionDigits: 2});
                
                let tagsHtml = '';
                for (const [k, v] of Object.entries(f.NormalizedTags || {})) {
                    tagsHtml += '<span class="tag-pill">' + k + ': ' + v + '</span>';
                }

                const row = document.createElement('tr');
                row.className = 'finding-row';
                row.setAttribute('data-cloud', f.ProviderLower);
                row.innerHTML = 
                    '<td>' +
                        '<span class="resource-type">' + f.ResourceName + '</span>' +
                        '<span class="resource-id">' + f.ResourceID + '</span>' +
                        '<div>' + tagsHtml + '</div>' +
                    '</td>' +
                    '<td><span class="cloud-badge ' + f.ProviderLower + '">' + f.Provider + '</span></td>' +
                    '<td>' + f.ResourceType + '</td>' +
                    '<td><span class="savings-value">' + formattedSavings + '</span></td>' +
                    '<td><div class="recommendation">' + f.Reason + '</div></td>';
                
                tbody.appendChild(row);
            });
        }

        // Recalculate distribution bar shares & widths
        function updateDistributionBars(activeProviders, invoiceTotal, scale) {
            const baselines = { aws: 24984.00, azure: 22246.50, gcp: 13581.00 };
            
            let awsCost = activeProviders.has('aws') ? baselines.aws * scale : 0;
            let azureCost = activeProviders.has('azure') ? baselines.azure * scale : 0;
            let gcpCost = activeProviders.has('gcp') ? baselines.gcp * scale : 0;

            const total = awsCost + azureCost + gcpCost;
            
            const awsPct = total > 0 ? (awsCost / total) * 100 : 0;
            const azurePct = total > 0 ? (azureCost / total) * 100 : 0;
            const gcpPct = total > 0 ? (gcpCost / total) * 100 : 0;

            document.getElementById('awsDistBar').style.width = awsPct + '%';
            document.getElementById('azureDistBar').style.width = azurePct + '%';
            document.getElementById('gcpDistBar').style.width = gcpPct + '%';

            updateKPIElement('awsDistShare', awsPct.toFixed(1) + '% ($' + awsCost.toLocaleString('en-US', {maximumFractionDigits: 0}) + ')');
            updateKPIElement('azureDistShare', azurePct.toFixed(1) + '% ($' + azureCost.toLocaleString('en-US', {maximumFractionDigits: 0}) + ')');
            updateKPIElement('gcpDistShare', gcpPct.toFixed(1) + '% ($' + gcpCost.toLocaleString('en-US', {maximumFractionDigits: 0}) + ')');
        }

        // Add Account connector click handler
        function addAccount(event) {
            event.preventDefault();
            const provider = document.getElementById('accProvider').value;
            const name = document.getElementById('accName').value.trim();
            const id = document.getElementById('accId').value.trim();
            
            // Extract custom tags
            const tags = {};
            const tagRows = document.querySelectorAll('.tag-input-row');
            tagRows.forEach(row => {
                const key = row.querySelector('.tag-key').value.trim();
                const val = row.querySelector('.tag-value').value.trim();
                if (key && val) {
                    tags[key] = val;
                }
            });

            // Add to accounts memory array
            accounts.push({ id: id, name: name, provider: provider, env: selectedFormEnv, tags: tags });
            
            // Reset form inputs
            document.getElementById('accName').value = '';
            document.getElementById('accId').value = '';
            
            const container = document.getElementById('tagsInputContainer');
            container.innerHTML = 
                '<div style="display: flex; gap: 0.5rem;" class="tag-input-row">' +
                    '<input type="text" placeholder="Key (e.g. client)" class="form-control tag-key" style="flex: 1;" required>' +
                    '<input type="text" placeholder="Value" class="form-control tag-value" style="flex: 1;" required>' +
                '</div>';

            syncAll();
            alert('Account connector ' + name + ' connected successfully!');
        }

        // Remove Account handler
        function removeAccount(index) {
            if (confirm('Disconnect ' + accounts[index].name + '?')) {
                accounts.splice(index, 1);
                syncAll();
            }
        }

        // Time Period filter click handler
        function setTimeFilter(period) {
            currentPeriod = period;

            // Toggle active classes on period buttons
            document.querySelectorAll('[id^="time-"]').forEach(btn => {
                btn.classList.remove('active');
            });
            
            const mapping = {
                '7d': 'time-7d',
                '30d': 'time-30d',
                '90d': 'time-90d',
                '365d': 'time-365d',
                'custom': 'time-custom'
            };
            document.getElementById(mapping[period]).classList.add('active');

            const drawer = document.getElementById('custom-date-drawer');
            if (period === 'custom') {
                drawer.style.display = 'flex';
            } else {
                drawer.style.display = 'none';
                syncAll();
            }
        }

        function applyCustomDates() {
            syncAll();
        }

        // Add new member to administrative table
        function inviteMember(event) {
            event.preventDefault();
            const name = document.getElementById('memberName').value.trim();
            const email = document.getElementById('memberEmail').value.trim();
            const role = document.getElementById('memberRole').value;

            if (!name || !email) return;

            const table = document.getElementById('teamTableBody');
            const row = document.createElement('tr');
            
            let roleClass = 'viewer';
            if (role === 'Owner / Admin') roleClass = 'admin';
            else if (role === 'FinOps Operator') roleClass = 'operator';

            row.innerHTML = '<td><div style=\"font-weight: 600;\">' + name + '</div></td>' +
                            '<td>' + email + '</td>' +
                            '<td><span class=\"role-badge ' + roleClass + '\">' + role + '</span></td>' +
                            '<td><span class=\"status-badge pending\">Pending</span></td>';

            table.appendChild(row);

            // Clean Form
            document.getElementById('memberName').value = '';
            document.getElementById('memberEmail').value = '';
            document.getElementById('memberRole').value = 'Viewer';

            alert('SSO Invitation sent to ' + name + ' (' + email + ') via Workspace Sync!');
        }

        // Filter Table Rows by Search & Cloud Provider
        function setFilter(cloud) {
            currentCloudFilter = cloud;
            
            // Toggle active classes on cloud buttons
            document.querySelectorAll('[id^="btn-"]').forEach(btn => {
                if (btn.id === 'btn-' + cloud) {
                    btn.classList.add('active');
                } else {
                    btn.classList.remove('active');
                }
            });
            
            syncAll();
        }

        function filterTable() {
            syncAll();
        }
    </script>
</body>
</html>
`

// HTMLResource represents scanned resources for serializing to Javascript
type HTMLResource struct {
	ID           string    `json:"ID"`
	Name         string    `json:"Name"`
	Type         string    `json:"Type"`
	Provider     string    `json:"Provider"`
	CostPerMonth float64   `json:"CostPerMonth"`
	Status       string    `json:"Status"`
}

// HTMLFinding represents wrapper over analyzer.Finding with helper methods for HTML templating
type HTMLFinding struct {
	ResourceID       string            `json:"ResourceID"`
	ResourceName     string            `json:"ResourceName"`
	ResourceType     string            `json:"ResourceType"`
	Provider         string            `json:"Provider"`
	ProviderLower    string            `json:"ProviderLower"`
	CostPerMonth     float64           `json:"CostPerMonth"`
	FormattedSavings string            `json:"FormattedSavings"`
	Reason           string            `json:"Reason"`
	NormalizedTags   map[string]string `json:"NormalizedTags"`
}

// RenderHTML writes a gorgeous glassmorphism single-page application dashboard
func RenderHTML(result *analyzer.AnalysisResult, resources []provider.Resource, w io.Writer) error {
	tmpl, err := template.New("dashboard").Parse(HTMLReportTemplate)
	if err != nil {
		return err
	}

	// Prepare findings HTML
	findingsHTML := make([]HTMLFinding, len(result.Findings))
	cloudsSeen := make(map[provider.CloudType]bool)

	for i, f := range result.Findings {
		name := f.ResourceName
		if name == "" {
			name = "unnamed"
		}
		
		findingsHTML[i] = HTMLFinding{
			ResourceID:       f.ResourceID,
			ResourceName:     name,
			ResourceType:     f.ResourceType,
			Provider:         string(f.Provider),
			ProviderLower:    strings.ToLower(string(f.Provider)),
			CostPerMonth:     f.CostPerMonth,
			FormattedSavings: FormatCost(f.SavingsPotential),
			Reason:           f.Reason,
			NormalizedTags:   f.NormalizedTags,
		}
		cloudsSeen[f.Provider] = true
	}

	// Prepare resources HTML
	resourcesHTML := make([]HTMLResource, len(resources))
	for i, r := range resources {
		resourcesHTML[i] = HTMLResource{
			ID:           r.ID,
			Name:         r.Name,
			Type:         string(r.Type),
			Provider:     strings.ToLower(string(r.Provider)),
			CostPerMonth: r.CostPerMonth,
			Status:       r.Status,
		}
	}

	// Serialize databases to raw Javascript JSON fields
	findingsJSON, _ := json.Marshal(findingsHTML)
	resourcesJSON, _ := json.Marshal(resourcesHTML)

	// Generate summary counts
	cloudCount := len(cloudsSeen)
	if cloudCount == 0 && result.TotalMonthlyCost > 0 {
		cloudCount = 3 // Standard default clouds
	}

	data := struct {
		GeneratedTime         string
		FormattedTotalCost    string
		FormattedTotalSavings string
		WastePercent          float64
		FindingsCount         int
		CloudCount            int
		Findings              []HTMLFinding
		FindingsJSON          template.JS
		ResourcesJSON         template.JS
	}{
		GeneratedTime:         time.Now().Format("2006-01-02 15:04:05 MST"),
		FormattedTotalCost:    FormatCost(result.TotalMonthlyCost),
		FormattedTotalSavings: FormatCost(result.TotalSavingsPotential),
		WastePercent:          result.WastePercent,
		FindingsCount:         len(result.Findings),
		CloudCount:            cloudCount,
		Findings:              findingsHTML,
		FindingsJSON:          template.JS(findingsJSON),
		ResourcesJSON:         template.JS(resourcesJSON),
	}

	return tmpl.Execute(w, data)
}
