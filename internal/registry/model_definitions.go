// Package registry provides model definitions and lookup helpers for various AI providers.
// Static model metadata is stored in model_definitions_static_data.go.
package registry

import (
	"sort"
	"strings"
)

// GetStaticModelDefinitionsByChannel returns static model definitions for a given channel/provider.
// It returns nil when the channel is unknown.
//
// Supported channels:
//   - claude
//   - gemini
//   - vertex
//   - gemini-cli
//   - aistudio
//   - codex
//   - qwen
//   - iflow
//   - antigravity (returns static overrides only)
func GetStaticModelDefinitionsByChannel(channel string) []*ModelInfo {
	key := strings.ToLower(strings.TrimSpace(channel))
	switch key {
	case "claude":
		return GetClaudeModels()
	case "gemini":
		return GetGeminiModels()
	case "vertex":
		return GetGeminiVertexModels()
	case "gemini-cli":
		return GetGeminiCLIModels()
	case "aistudio":
		return GetAIStudioModels()
	case "codex":
		return GetOpenAIModels()
	case "qwen":
		return GetQwenModels()
	case "iflow":
		return GetIFlowModels()
	case "antigravity":
		cfg := GetAntigravityModelConfig()
		if len(cfg) == 0 {
			return nil
		}
		models := make([]*ModelInfo, 0, len(cfg))
		for modelID, entry := range cfg {
			if modelID == "" || entry == nil {
				continue
			}
			models = append(models, &ModelInfo{
				ID:                  modelID,
				Object:              "model",
				OwnedBy:             "antigravity",
				Type:                "antigravity",
				Thinking:            entry.Thinking,
				MaxCompletionTokens: entry.MaxCompletionTokens,
			})
		}
		sort.Slice(models, func(i, j int) bool {
			return strings.ToLower(models[i].ID) < strings.ToLower(models[j].ID)
		})
		return models
	default:
		return nil
	}
}

// LookupStaticModelInfo searches all static model definitions for a model by ID.
// Returns nil if no matching model is found.
func LookupStaticModelInfo(modelID string) *ModelInfo {
	if modelID == "" {
		return nil
	}

	allModels := [][]*ModelInfo{
		GetClaudeModels(),
		GetGeminiModels(),
		GetGeminiVertexModels(),
		GetGeminiCLIModels(),
		GetAIStudioModels(),
		GetOpenAIModels(),
		GetQwenModels(),
		GetIFlowModels(),
	}
	for _, models := range allModels {
		for _, m := range models {
			if m != nil && m.ID == modelID {
				return m
			}
		}
	}

	// Check Antigravity static config
	if cfg := GetAntigravityModelConfig()[modelID]; cfg != nil {
		return &ModelInfo{
			ID:                  modelID,
			Thinking:            cfg.Thinking,
			MaxCompletionTokens: cfg.MaxCompletionTokens,
		}
	}

	return nil
}

// GetGitHubCopilotModels returns the available models for GitHub Copilot.
// These models are available through the GitHub Copilot API at api.githubcopilot.com.
func GetGitHubCopilotModels() []*ModelInfo {
	now := int64(1768908139) // 2026-01-20
	return []*ModelInfo{
		{
			ID:                  "claude-haiku-4.5",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Claude Haiku 4.5",
			Description:         "Claude Haiku 4.5 via GitHub Copilot",
			ContextLength:       144000,
			MaxCompletionTokens: 16000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "claude-opus-4.5",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Claude Opus 4.5",
			Description:         "Claude Opus 4.5 via GitHub Copilot",
			ContextLength:       160000,
			MaxCompletionTokens: 16000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "claude-sonnet-4",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Claude Sonnet 4",
			Description:         "Claude Sonnet 4 via GitHub Copilot",
			ContextLength:       216000,
			MaxCompletionTokens: 16000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "claude-sonnet-4.5",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Claude Sonnet 4.5",
			Description:         "Claude Sonnet 4.5 via GitHub Copilot",
			ContextLength:       144000,
			MaxCompletionTokens: 16000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "gemini-2.5-pro",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Gemini 2.5 Pro",
			Description:         "Gemini 2.5 Pro via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 128, Max: 32768, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "gemini-3-flash-preview",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Gemini 3 Flash (Preview)",
			Description:         "Gemini 3 Flash (Preview) via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 256, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "gemini-3-pro-preview",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Gemini 3 Pro (Preview)",
			Description:         "Gemini 3 Pro (Preview) via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Min: 258, Max: 32000, ZeroAllowed: false, DynamicAllowed: false},
		},
		{
			ID:                  "gpt-3.5-turbo",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 3.5 Turbo",
			Description:         "GPT 3.5 Turbo via GitHub Copilot",
			ContextLength:       16384,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-3.5-turbo-0613",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 3.5 Turbo",
			Description:         "GPT 3.5 Turbo via GitHub Copilot",
			ContextLength:       16384,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 4",
			Description:         "GPT 4 via GitHub Copilot",
			ContextLength:       32768,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 4",
			Description:         "GPT 4 via GitHub Copilot",
			ContextLength:       32768,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4-0125-preview",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 4 Turbo",
			Description:         "GPT 4 Turbo via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4-0613",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT 4",
			Description:         "GPT 4 via GitHub Copilot",
			ContextLength:       32768,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4-o-preview",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4-o-preview",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4.1",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4.1",
			Description:         "GPT-4.1 via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4.1-2025-04-14",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4.1",
			Description:         "GPT-4.1 via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o-2024-05-13",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o-2024-08-06",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o-2024-11-20",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o",
			Description:         "GPT-4o via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o-mini",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o mini",
			Description:         "GPT-4o mini via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-4o-mini-2024-07-18",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-4o mini",
			Description:         "GPT-4o mini via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 4096,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5",
			Description:         "GPT-5 via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/chat/completions", "/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5-codex",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5-Codex (Preview)",
			Description:         "GPT-5-Codex (Preview) via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5-mini",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5 mini",
			Description:         "GPT-5 mini via GitHub Copilot",
			ContextLength:       264000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.1",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.1",
			Description:         "GPT-5.1 via GitHub Copilot",
			ContextLength:       264000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions", "/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.1-codex",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.1-Codex",
			Description:         "GPT-5.1-Codex via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.1-codex-max",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.1-Codex-Max",
			Description:         "GPT-5.1-Codex-Max via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.1-codex-mini",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.1-Codex-Mini",
			Description:         "GPT-5.1-Codex-Mini via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.2",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.2",
			Description:         "GPT-5.2 via GitHub Copilot",
			ContextLength:       264000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions", "/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "gpt-5.2-codex",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "GPT-5.2-Codex",
			Description:         "GPT-5.2-Codex via GitHub Copilot",
			ContextLength:       400000,
			MaxCompletionTokens: 128000,
			SupportedEndpoints:  []string{"/responses"},
			Thinking:            &ThinkingSupport{Levels: []string{"minimal", "low", "medium", "high"}},
		},
		{
			ID:                  "grok-code-fast-1",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Grok Code Fast 1",
			Description:         "Grok Code Fast 1 via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 64000,
			SupportedEndpoints:  []string{"/chat/completions"},
		},
		{
			ID:                  "text-embedding-3-small",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Embedding V3 small",
			Description:         "Embedding V3 small via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
		},
		{
			ID:                  "text-embedding-3-small-inference",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Embedding V3 small (Inference)",
			Description:         "Embedding V3 small (Inference) via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
		},
		{
			ID:                  "text-embedding-ada-002",
			Object:              "model",
			Created:             now,
			OwnedBy:             "github-copilot",
			Type:                "github-copilot",
			DisplayName:         "Embedding V2 Ada",
			Description:         "Embedding V2 Ada via GitHub Copilot",
			ContextLength:       128000,
			MaxCompletionTokens: 16384,
		},
	}
}

// GetKiroModels returns the Kiro (AWS CodeWhisperer) model definitions
func GetKiroModels() []*ModelInfo {
	return []*ModelInfo{
		// --- Base Models ---
		{
			ID:                  "kiro-auto",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Auto",
			Description:         "Automatic model selection by Kiro",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-opus-4-5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Opus 4.5",
			Description:         "Claude Opus 4.5 via Kiro (2.2x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-sonnet-4-5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Sonnet 4.5",
			Description:         "Claude Sonnet 4.5 via Kiro (1.3x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-sonnet-4",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Sonnet 4",
			Description:         "Claude Sonnet 4 via Kiro (1.3x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-haiku-4-5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Haiku 4.5",
			Description:         "Claude Haiku 4.5 via Kiro (0.4x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		// --- Agentic Variants (Optimized for coding agents with chunked writes) ---
		{
			ID:                  "kiro-claude-opus-4-5-agentic",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Opus 4.5 (Agentic)",
			Description:         "Claude Opus 4.5 optimized for coding agents (chunked writes)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-sonnet-4-5-agentic",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Sonnet 4.5 (Agentic)",
			Description:         "Claude Sonnet 4.5 optimized for coding agents (chunked writes)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-sonnet-4-agentic",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Sonnet 4 (Agentic)",
			Description:         "Claude Sonnet 4 optimized for coding agents (chunked writes)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
		{
			ID:                  "kiro-claude-haiku-4-5-agentic",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Kiro Claude Haiku 4.5 (Agentic)",
			Description:         "Claude Haiku 4.5 optimized for coding agents (chunked writes)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
			Thinking:            &ThinkingSupport{Min: 1024, Max: 32000, ZeroAllowed: true, DynamicAllowed: true},
		},
	}
}

// GetAmazonQModels returns the Amazon Q (AWS CodeWhisperer) model definitions.
// These models use the same API as Kiro and share the same executor.
func GetAmazonQModels() []*ModelInfo {
	return []*ModelInfo{
		{
			ID:                  "amazonq-auto",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro", // Uses Kiro executor - same API
			DisplayName:         "Amazon Q Auto",
			Description:         "Automatic model selection by Amazon Q",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
		},
		{
			ID:                  "amazonq-claude-opus-4.5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Amazon Q Claude Opus 4.5",
			Description:         "Claude Opus 4.5 via Amazon Q (2.2x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
		},
		{
			ID:                  "amazonq-claude-sonnet-4.5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Amazon Q Claude Sonnet 4.5",
			Description:         "Claude Sonnet 4.5 via Amazon Q (1.3x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
		},
		{
			ID:                  "amazonq-claude-sonnet-4",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Amazon Q Claude Sonnet 4",
			Description:         "Claude Sonnet 4 via Amazon Q (1.3x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
		},
		{
			ID:                  "amazonq-claude-haiku-4.5",
			Object:              "model",
			Created:             1732752000,
			OwnedBy:             "aws",
			Type:                "kiro",
			DisplayName:         "Amazon Q Claude Haiku 4.5",
			Description:         "Claude Haiku 4.5 via Amazon Q (0.4x credit)",
			ContextLength:       200000,
			MaxCompletionTokens: 64000,
		},
	}
}
