package management

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	githubAPIBaseURL      = "https://api.github.com"
	copilotUserEndpoint   = "/copilot_internal/user"
	copilotUserAgent      = "GitHubCopilotChat/0.26.7"
	copilotEditorVersion  = "vscode/1.100.0"
	copilotPluginVersion  = "copilot-chat/0.26.7"
	copilotAPIVersion     = "2025-04-01"
)

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

// GetCopilotQuota handles the GET request to fetch Copilot quota.
// It expects an 'auth_id' query parameter.
func (h *Handler) GetCopilotQuota(c *gin.Context) {
	authID := c.Query("auth_id")
	if authID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "auth_id is required"})
		return
	}

	if h.authManager == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "auth manager not available"})
		return
	}

	auth, ok := h.authManager.GetByID(authID)
	if !ok || auth == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "auth not found"})
		return
	}

	// Try to get token from metadata
	var token string
	if auth.Metadata != nil {
		if t, ok := auth.Metadata["access_token"].(string); ok {
			token = t
		}
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no access token found in auth"})
		return
	}

	usage, err := fetchCopilotUsage(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch usage: %v", err)})
		return
	}

	c.JSON(http.StatusOK, usage)
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
