package management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	kiroauth "github.com/router-for-me/CLIProxyAPI/v6/internal/auth/kiro"
)

// GetKiroQuota handles the GET request to fetch Kiro quota.
// It expects an 'auth_id' query parameter.
func (h *Handler) GetKiroQuota(c *gin.Context) {
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

	// Reconstruct KiroTokenData from auth.Metadata
	tokenData := &kiroauth.KiroTokenData{}
	metadataJSON, err := json.Marshal(auth.Metadata)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to marshal metadata: %v", err)})
		return
	}
	if err := json.Unmarshal(metadataJSON, tokenData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to unmarshal token data: %v", err)})
		return
	}

	// Ensure AccessToken is present
	if tokenData.AccessToken == "" {
		// Fallback: try to get access_token directly if json unmarshal didn't work as expected for some reason
		if t, ok := auth.Metadata["access_token"].(string); ok {
			tokenData.AccessToken = t
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "access_token not found in auth data"})
			return
		}
	}
    
    // Also ensure other fields are populated if possible
    if tokenData.Region == "" {
        if r, ok := auth.Metadata["region"].(string); ok {
            tokenData.Region = r
        }
    }

	kAuth := kiroauth.NewKiroAuth(h.cfg)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	usage, err := kAuth.GetUsageLimits(ctx, tokenData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch usage: %v", err)})
		return
	}

	c.JSON(http.StatusOK, usage)
}
