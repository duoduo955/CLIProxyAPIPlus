// Package main provides a tool to automatically update GitHub Copilot model definitions
// in the model_definitions.go file by fetching the latest models from the Copilot API.
//
// Usage:
//
//	go run tools/update_copilot_models/main.go
//
// The tool will:
// 1. Read GitHub Copilot auth files from the auth directory
// 2. Exchange the GitHub access token for a Copilot API token
// 3. Fetch the latest model list from https://api.githubcopilot.com/models
// 4. Generate the updated GetGitHubCopilotModels() function
// 5. Update internal/registry/model_definitions.go in place
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	copilotTokenURL  = "https://api.github.com/copilot_internal/v2/token"
	copilotModelsURL = "https://api.githubcopilot.com/models"
	targetFile       = "internal/registry/model_definitions.go"

	// HTTP header values
	copilotUserAgent     = "GithubCopilot/1.0"
	copilotEditorVersion = "vscode/1.100.0"
	copilotPluginVersion = "copilot/1.300.0"
	copilotIntegrationID = "vscode-chat"
)

// AuthFile represents the structure of a Copilot auth JSON file
type AuthFile struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token"`
	Username    string `json:"username"`
}

// CopilotAPIToken represents the token exchange response
type CopilotAPIToken struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

// CopilotModelsResponse represents the API response from /models endpoint
type CopilotModelsResponse struct {
	Data   []CopilotModel `json:"data"`
	Object string         `json:"object"`
}

// CopilotModel represents a single model from the API
type CopilotModel struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Object             string              `json:"object"`
	Version            string              `json:"version"`
	Vendor             string              `json:"vendor"`
	Preview            bool                `json:"preview"`
	Capabilities       CopilotCapabilities `json:"capabilities"`
	SupportedEndpoints []string            `json:"supported_endpoints"`
}

// CopilotCapabilities represents model capabilities
type CopilotCapabilities struct {
	Family    string        `json:"family"`
	Type      string        `json:"type"`
	Tokenizer string        `json:"tokenizer"`
	Limits    CopilotLimits `json:"limits"`
	Supports  CopilotSupports `json:"supports"`
}

// CopilotSupports represents what features the model supports
type CopilotSupports struct {
	MaxThinkingBudget  int  `json:"max_thinking_budget"`
	MinThinkingBudget  int  `json:"min_thinking_budget"`
	ParallelToolCalls  bool `json:"parallel_tool_calls"`
	Streaming          bool `json:"streaming"`
	StructuredOutputs  bool `json:"structured_outputs"`
	ToolCalls          bool `json:"tool_calls"`
	Vision             bool `json:"vision"`
}

// CopilotLimits represents model token limits
type CopilotLimits struct {
	MaxContextWindowTokens int `json:"max_context_window_tokens"`
	MaxOutputTokens        int `json:"max_output_tokens"`
	MaxPromptTokens        int `json:"max_prompt_tokens"`
}

func main() {
	authDir := flag.String("auth-dir", "", "Auth directory path (default: ~/.cli-proxy-api)")
	dryRun := flag.Bool("dry-run", false, "Print generated code without updating file")
	outputFile := flag.String("output", "", "Save raw API response to file (e.g., copilot_models.json)")
	flag.Parse()

	// Determine auth directory
	authPath := *authDir
	if authPath == "" {
		authPath = getDefaultAuthDir()
	}
	authPath = expandHome(authPath)

	fmt.Printf("Looking for Copilot auth files in: %s\n", authPath)

	// Find and read Copilot auth file
	accessToken, username, err := findCopilotAuthToken(authPath)
	if err != nil {
		fmt.Printf("Error finding auth token: %v\n", err)
		fmt.Println("\nPlease ensure you have logged in with GitHub Copilot using:")
		fmt.Println("  CLIProxyAPI github-copilot-login")
		os.Exit(1)
	}

	fmt.Printf("Found Copilot auth for user: %s\n", username)

	// Exchange GitHub token for Copilot API token
	fmt.Println("Exchanging token for Copilot API access...")
	apiToken, err := getCopilotAPIToken(accessToken)
	if err != nil {
		fmt.Printf("Error getting Copilot API token: %v\n", err)
		os.Exit(1)
	}

	// Fetch models
	fmt.Println("Fetching models from GitHub Copilot API...")
	models, rawJSON, err := fetchCopilotModels(apiToken.Token)
	if err != nil {
		fmt.Printf("Error fetching models: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Fetched %d models from API\n", len(models))

	// Save raw API response to file if requested
	if *outputFile != "" {
		var prettyJSON map[string]interface{}
		var formatted []byte
		if err := json.Unmarshal(rawJSON, &prettyJSON); err == nil {
			formatted, _ = json.MarshalIndent(prettyJSON, "", "  ")
		} else {
			formatted = rawJSON
		}

		if err := os.WriteFile(*outputFile, formatted, 0644); err != nil {
			fmt.Printf("Warning: Failed to save output to %s: %v\n", *outputFile, err)
		} else {
			fmt.Printf("Raw API response saved to: %s\n", *outputFile)
		}
	}

	// Print raw API response for inspection (only if no output file specified)
	if *outputFile == "" {
		fmt.Println("\n=== RAW API RESPONSE ===")
		var prettyJSON map[string]interface{}
		if err := json.Unmarshal(rawJSON, &prettyJSON); err == nil {
			formatted, _ := json.MarshalIndent(prettyJSON, "", "  ")
			fmt.Println(string(formatted))
		} else {
			fmt.Println(string(rawJSON))
		}
		fmt.Println("=== END RAW API RESPONSE ===\n")
	}

	// Print model list with detailed info
	fmt.Println("\nModels found:")
	for _, m := range models {
		fmt.Printf("  - %s (%s)\n", m.ID, m.Vendor)
		fmt.Printf("    Family: %s, Type: %s\n", m.Capabilities.Family, m.Capabilities.Type)
		fmt.Printf("    Context: %d, Output: %d\n",
			m.Capabilities.Limits.MaxContextWindowTokens,
			m.Capabilities.Limits.MaxOutputTokens)
	}

	// Generate the Go code
	code := generateGoCode(models)

	if *dryRun {
		fmt.Println("\n--- Generated Code ---")
		fmt.Println(code)
		fmt.Println("--- End of Generated Code ---")
		return
	}

	// Find project root
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Printf("Error finding project root: %v\n", err)
		os.Exit(1)
	}

	targetPath := filepath.Join(projectRoot, targetFile)
	err = updateModelDefinitions(targetPath, code)
	if err != nil {
		fmt.Printf("Error updating file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\nSuccessfully updated %s with %d models\n", targetFile, len(models))
}

func getDefaultAuthDir() string {
	return "~/.cli-proxy-api"
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			// Fallback for Windows
			if runtime.GOOS == "windows" {
				home = os.Getenv("USERPROFILE")
			}
		}
		if home != "" {
			if path == "~" {
				return home
			}
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

func findCopilotAuthToken(authDir string) (accessToken, username string, err error) {
	entries, err := os.ReadDir(authDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to read auth directory: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}

		fullPath := filepath.Join(authDir, name)
		data, errRead := os.ReadFile(fullPath)
		if errRead != nil {
			continue
		}

		var auth AuthFile
		if errUnmarshal := json.Unmarshal(data, &auth); errUnmarshal != nil {
			continue
		}

		// Check if this is a GitHub Copilot auth file
		if strings.ToLower(auth.Type) == "github-copilot" && auth.AccessToken != "" {
			return auth.AccessToken, auth.Username, nil
		}
	}

	return "", "", fmt.Errorf("no GitHub Copilot auth file found in %s", authDir)
}

func getCopilotAPIToken(githubAccessToken string) (*CopilotAPIToken, error) {
	req, err := http.NewRequest("GET", copilotTokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+githubAccessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", copilotUserAgent)
	req.Header.Set("Editor-Version", copilotEditorVersion)
	req.Header.Set("Editor-Plugin-Version", copilotPluginVersion)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var apiToken CopilotAPIToken
	if err := json.Unmarshal(body, &apiToken); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiToken.Token == "" {
		return nil, fmt.Errorf("received empty API token")
	}

	return &apiToken, nil
}

func fetchCopilotModels(apiToken string) ([]CopilotModel, []byte, error) {
	req, err := http.NewRequest("GET", copilotModelsURL, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", copilotUserAgent)
	req.Header.Set("Editor-Version", copilotEditorVersion)
	req.Header.Set("Editor-Plugin-Version", copilotPluginVersion)
	req.Header.Set("Copilot-Integration-Id", copilotIntegrationID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response: %w", err)
	}

	var modelsResp CopilotModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, body, fmt.Errorf("failed to parse response: %w", err)
	}

	return modelsResp.Data, body, nil
}

func generateGoCode(models []CopilotModel) string {
	// Sort models by ID for consistent output
	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	now := time.Now().Unix()
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("func GetGitHubCopilotModels() []*ModelInfo {\n"))
	sb.WriteString(fmt.Sprintf("\tnow := int64(%d) // %s\n", now, time.Now().Format("2006-01-02")))
	sb.WriteString("\treturn []*ModelInfo{\n")

	for _, m := range models {
		if m.ID == "" {
			continue
		}

		displayName := m.Name
		if displayName == "" {
			displayName = m.ID
		}

		vendor := m.Vendor
		if vendor == "" {
			vendor = "github-copilot"
		}

		// Generate description based on vendor
		description := fmt.Sprintf("%s via GitHub Copilot", displayName)

		contextLength := m.Capabilities.Limits.MaxContextWindowTokens
		maxOutputTokens := m.Capabilities.Limits.MaxOutputTokens

		// Use defaults if not specified
		if contextLength == 0 {
			contextLength = 128000
		}
		if maxOutputTokens == 0 {
			maxOutputTokens = 16384
		}

		// Determine supported endpoints
		endpoints := determineSupportedEndpoints(m)

		sb.WriteString("\t\t{\n")
		sb.WriteString(fmt.Sprintf("\t\t\tID:                  %q,\n", m.ID))
		sb.WriteString("\t\t\tObject:              \"model\",\n")
		sb.WriteString("\t\t\tCreated:             now,\n")
		sb.WriteString("\t\t\tOwnedBy:             \"github-copilot\",\n")
		sb.WriteString("\t\t\tType:                \"github-copilot\",\n")
		sb.WriteString(fmt.Sprintf("\t\t\tDisplayName:         %q,\n", displayName))
		sb.WriteString(fmt.Sprintf("\t\t\tDescription:         %q,\n", description))
		sb.WriteString(fmt.Sprintf("\t\t\tContextLength:       %d,\n", contextLength))
		sb.WriteString(fmt.Sprintf("\t\t\tMaxCompletionTokens: %d,\n", maxOutputTokens))

		// Add SupportedEndpoints if available
		if len(endpoints) > 0 {
			sb.WriteString(fmt.Sprintf("\t\t\tSupportedEndpoints:  %s,\n", formatStringSlice(endpoints)))
		}

		// Add Thinking support if available
		if thinkingStr := generateThinkingSupport(m); thinkingStr != "" {
			sb.WriteString(fmt.Sprintf("\t\t\tThinking:            %s,\n", thinkingStr))
		}

		sb.WriteString("\t\t},\n")
	}

	sb.WriteString("\t}\n")
	sb.WriteString("}\n")

	return sb.String()
}

// determineSupportedEndpoints determines which API endpoints a model supports
func determineSupportedEndpoints(model CopilotModel) []string {
	// If API provides supported_endpoints, use it
	if len(model.SupportedEndpoints) > 0 {
		return model.SupportedEndpoints
	}

	modelLower := strings.ToLower(model.ID)

	// Embedding models don't support chat/responses endpoints
	if strings.Contains(modelLower, "embedding") {
		return []string{}
	}

	// Codex models only support /responses
	if strings.Contains(modelLower, "-codex") {
		return []string{"/responses"}
	}

	// Default: support /chat/completions (safest option)
	return []string{"/chat/completions"}
}

// generateThinkingSupport generates the ThinkingSupport configuration
func generateThinkingSupport(model CopilotModel) string {
	maxBudget := model.Capabilities.Supports.MaxThinkingBudget
	minBudget := model.Capabilities.Supports.MinThinkingBudget

	// If model has thinking budget info, use it
	if maxBudget > 0 {
		return fmt.Sprintf("&ThinkingSupport{Min: %d, Max: %d, ZeroAllowed: false, DynamicAllowed: false}",
			minBudget, maxBudget)
	}

	// Check if it's a GPT model that supports level-based thinking
	modelLower := strings.ToLower(model.ID)
	if strings.HasPrefix(modelLower, "gpt-") || strings.HasPrefix(modelLower, "o1") || strings.HasPrefix(modelLower, "o3") {
		return "&ThinkingSupport{Levels: []string{\"minimal\", \"low\", \"medium\", \"high\"}}"
	}

	return ""
}

// formatStringSlice formats a string slice as Go code
func formatStringSlice(slice []string) string {
	if len(slice) == 0 {
		return "[]string{}"
	}

	quoted := make([]string, len(slice))
	for i, s := range slice {
		quoted[i] = fmt.Sprintf("%q", s)
	}

	return fmt.Sprintf("[]string{%s}", strings.Join(quoted, ", "))
}

func findProjectRoot() (string, error) {
	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	fmt.Printf("Starting search from: %s\n", dir)

	// Try to find go.mod file by walking up the directory tree
	searchPath := dir
	for {
		goModPath := filepath.Join(searchPath, "go.mod")
		fmt.Printf("Checking: %s\n", goModPath)

		if _, err := os.Stat(goModPath); err == nil {
			// Found go.mod, this is the project root
			fmt.Printf("Found project root: %s\n", searchPath)
			return searchPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(searchPath)
		if parent == searchPath {
			// Reached filesystem root without finding go.mod
			break
		}
		searchPath = parent
	}

	return "", fmt.Errorf("could not find project root (go.mod not found), searched from: %s", dir)
}

func updateModelDefinitions(filePath string, newCode string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Match the entire GetGitHubCopilotModels function
	// Pattern: from "func GetGitHubCopilotModels()" to the closing "}\n" before next function or EOF
	pattern := regexp.MustCompile(`(?s)func GetGitHubCopilotModels\(\) \[\]\*ModelInfo \{.*?\n\}\n`)

	if !pattern.MatchString(string(content)) {
		return fmt.Errorf("could not find GetGitHubCopilotModels function in file")
	}

	newContent := pattern.ReplaceAllString(string(content), newCode)

	err = os.WriteFile(filePath, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
