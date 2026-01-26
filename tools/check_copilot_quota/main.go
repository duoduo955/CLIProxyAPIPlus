package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	githubAPIBaseURL      = "https://api.github.com"
	copilotUserEndpoint   = "/copilot_internal/user"
	copilotUserAgent      = "GitHubCopilotChat/0.26.7"

copilotEditorVersion  = "vscode/1.100.0"
copilotPluginVersion  = "copilot-chat/0.26.7"
copilotAPIVersion     = "2025-04-01"
)

// TokenFile represents the structure of the auth file
type TokenFile struct {
	AccessToken string `json:"access_token"`
}

// QuotaDetail represents the quota information for a single quota type.
type QuotaDetail struct {
	Entitlement      float64 `json:"entitlement"`
	OverageCount     float64 `json:"overage_count"`
	OveragePermitted bool    `json:"overage_permitted"`
	PercentRemaining float64 `json:"percent_remaining"`
	QuotaID          string  `json:"quota_id"`
	QuotaRemaining   float64 `json:"quota_remaining"`
	Remaining        float64 `json:"remaining"`
	Unlimited        bool    `json:"unlimited"`
}

// QuotaSnapshots contains quota information for different interaction types.
type QuotaSnapshots struct {
	Chat                QuotaDetail `json:"chat"`
	Completions         QuotaDetail `json:"completions"`
	PremiumInteractions QuotaDetail `json:"premium_interactions"`
}

// CopilotUsageResponse represents the response from the Copilot usage endpoint.
type CopilotUsageResponse struct {
	AccessTypeSKU           string         `json:"access_type_sku"`
	AnalyticsTrackingID     string         `json:"analytics_tracking_id"`
	AssignedDate            string         `json:"assigned_date"`
	CanSignupForLimited     bool           `json:"can_signup_for_limited"`
	ChatEnabled             bool           `json:"chat_enabled"`
	CopilotPlan             string         `json:"copilot_plan"`
	OrganizationLoginList   []interface{}  `json:"organization_login_list"`
	OrganizationList        []interface{}  `json:"organization_list"`
	QuotaResetDate          string         `json:"quota_reset_date"`
	QuotaSnapshots          QuotaSnapshots `json:"quota_snapshots"`
}

func main() {
	tokenFlag := flag.String("token", "", "GitHub Copilot Access Token (ghu_... or gho_...)")
	fileFlag := flag.String("file", "", "Path to the auth JSON file (e.g., github-copilot-xxx.json)")
	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	finalToken := *tokenFlag
	filePath := *fileFlag

	// 1. If no token/file provided, try to find a file automatically by reading config.yaml
	if finalToken == "" && filePath == "" {
		// Try to read config.yaml to get auth-dir
		authDir := getAuthDirFromConfig()
		if authDir != "" {
			if !*jsonOutput {
				fmt.Printf("Found auth-dir from config: %s\n", authDir)
			}
			foundFile, err := findTokenFileInDir(authDir)
			if err == nil && foundFile != "" {
				filePath = foundFile
				if !*jsonOutput {
					fmt.Printf("Auto-detected token file: %s\n", filePath)
				}
			}
		} else {
			// Fallback to default paths if config read fails
			foundFile, err := findDefaultTokenFile()
			if err == nil && foundFile != "" {
				filePath = foundFile
				if !*jsonOutput {
					fmt.Printf("Auto-detected token file (fallback): %s\n", filePath)
				}
			}
		}
	}

	// 2. Try reading from file if we have a file path
	if finalToken == "" && filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			printError(*jsonOutput, fmt.Sprintf("Error reading file %s: %v", filePath, err))
			os.Exit(1)
		}
		var tf TokenFile
		if err := json.Unmarshal(data, &tf); err != nil {
			printError(*jsonOutput, fmt.Sprintf("Error parsing JSON file: %v", err))
			os.Exit(1)
		}
		if tf.AccessToken == "" {
			printError(*jsonOutput, "'access_token' field not found in the file")
			os.Exit(1)
		}
		finalToken = tf.AccessToken
	}

	// 3. Try Environment Variable
	if finalToken == "" {
		finalToken = os.Getenv("GITHUB_COPILOT_TOKEN")
	}

	// 4. Final check
	if finalToken == "" {
		if *jsonOutput {
			fmt.Println(`{"error": "No token found. Provide -file, -token, or set GITHUB_COPILOT_TOKEN."}`) 
		} else {
			fmt.Println("Error: No token found.")
			fmt.Println("Could not find 'auth-dir' in config.yaml or 'github-copilot*.json' in default paths.")
			fmt.Println("Usage:")
			fmt.Println("  go run main.go [-file path/to/file.json] [-token ghu_...] [-json]")
		}
		os.Exit(1)
	}

	if !*jsonOutput {
		fmt.Println("Fetching Copilot usage data...")
	}

	usage, err := fetchCopilotUsage(finalToken)
	if err != nil {
		printError(*jsonOutput, fmt.Sprintf("Error fetching usage: %v", err))
		os.Exit(1)
	}

	if *jsonOutput {
		b, _ := json.MarshalIndent(usage, "", "  ")
		fmt.Println(string(b))
	} else {
		printReport(usage)
	}
}

// getAuthDirFromConfig tries to read 'auth-dir' from config.yaml in current or parent dir
func getAuthDirFromConfig() string {
	searchFiles := []string{"config.yaml", "../config.yaml", "CLIProxyAPIPlus-main/config.yaml"}
	
	for _, file := range searchFiles {
		if _, err := os.Stat(file); err == nil {
			dir := parseAuthDir(file)
			if dir != "" {
				return dir
			}
		}
	}
	return ""
}

// parseAuthDir reads the yaml file line by line to find "auth-dir: value"
func parseAuthDir(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	// regex to match "auth-dir: value" or "auth-dir: 'value'" or "auth-dir: "value""
	// Handling simple cases without full YAML parser dependency
	re := regexp.MustCompile(`^\s*auth-dir:\s*["']?([^"'\s]+)["']?`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)
		if len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

// findDefaultTokenFile searches for github-copilot*.json files in common directories
func findDefaultTokenFile() (string, error) {
	// Fallback paths if config.yaml search fails
	// ENSURE DOUBLE QUOTES ARE USED HERE
	searchPaths := []string{`.cli-proxy-api`, `auths`, `.`}
	for _, dir := range searchPaths {
		file, err := findTokenFileInDir(dir)
		if err == nil && file != "" {
			return file, nil
		}
	}
	return "", fmt.Errorf("no token file found")
}

func findTokenFileInDir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "github-copilot") && strings.HasSuffix(entry.Name(), ".json") {
			return filepath.Join(dir, entry.Name()), nil
		}
	}
	return "", nil
}

func printError(jsonMode bool, msg string) {
	if jsonMode {
		// Simple JSON string escape
		escaped := strings.ReplaceAll(msg, `"`, `\"`)
		fmt.Printf(`{"error": "%s"}`, escaped)
		fmt.Println()
	} else {
		fmt.Println(msg)
	}
}

func fetchCopilotUsage(token string) (*CopilotUsageResponse, error) {
	url := githubAPIBaseURL + copilotUserEndpoint

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "token "+token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", copilotUserAgent)
	req.Header.Set("Editor-Version", copilotEditorVersion)
	req.Header.Set("Editor-Plugin-Version", copilotPluginVersion)
	req.Header.Set("X-GitHub-Api-Version", copilotAPIVersion)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var usage CopilotUsageResponse
	if err := json.Unmarshal(body, &usage); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &usage, nil
}

func printReport(data *CopilotUsageResponse) {
	fmt.Println("\n==================================================")
	fmt.Println(" GITHUB COPILOT USAGE REPORT")
	fmt.Println("==================================================")

	fmt.Printf("Plan:        %s\n", data.CopilotPlan)
	fmt.Printf("Chat:        %v\n", data.ChatEnabled)
	fmt.Printf("Reset Date:  %s\n", data.QuotaResetDate)
	fmt.Println("--------------------------------------------------")

	// Helper to print quota
	printQuota := func(name string, q QuotaDetail) {
		used := q.Entitlement - q.Remaining
		percentUsed := 0.0
		
		unlimitedStr := ""
		if q.Unlimited {
			unlimitedStr = " (Unlimited)"
		}

		fmt.Printf("%s%s:\n", name, unlimitedStr)
		if q.Entitlement > 0 {
			percentUsed = (used / q.Entitlement) * 100
			fmt.Printf("  Used:  %.0f / %.0f\n", used, q.Entitlement)
			fmt.Printf("  Stats: %.1f%% used, %.1f%% remaining\n", percentUsed, q.PercentRemaining)
		} else {
			// If entitlement is 0, just show remaining if it has value, or N/A
			fmt.Printf("  Remaining: %.0f\n", q.Remaining)
		}
		fmt.Println("--------------------------------------------------")
	}

	printQuota("Premium Interactions (Models)", data.QuotaSnapshots.PremiumInteractions)
	printQuota("Standard Chat", data.QuotaSnapshots.Chat)
	printQuota("Code Completions", data.QuotaSnapshots.Completions)

	fmt.Println("==================================================")
}
