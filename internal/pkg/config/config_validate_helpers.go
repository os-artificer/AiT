/**
 * Copyright 2026 AiT authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
**/

package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/platform"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func configNotFoundErr(path string) error {
	return errhint.WithTryMsg(
		"config file not found: "+path,
		fmt.Sprintf("Create %s before validating.", platform.ConfigFileName),
		toolname.CLI("config init"),
	)
}

func validateSemantics(cfg *Config) []ValidationIssue {
	if cfg == nil {
		return nil
	}

	var issues []ValidationIssue
	issues = append(issues, validateSession(cfg)...)
	issues = append(issues, validateAI(cfg)...)
	issues = append(issues, legacyDeepseekWarning(cfg)...)
	return issues
}

func validateSession(cfg *Config) []ValidationIssue {
	var issues []ValidationIssue

	if cfg.Session.AuthTTL != "" {
		if _, err := time.ParseDuration(cfg.Session.AuthTTL); err != nil {
			issues = append(issues, ValidationIssue{
				Field:   keySessionAuthTTL,
				Message: fmt.Sprintf("invalid duration %q", cfg.Session.AuthTTL),
			})
		}
	}

	path := strings.ToLower(strings.TrimSpace(cfg.Session.RemotePrompt.Path))
	if path != "" && path != RemotePromptPathShort && path != RemotePromptPathAbsolute {
		issues = append(issues, ValidationIssue{
			Field:   "session.remote_prompt.path",
			Message: fmt.Sprintf("must be %q or %q", RemotePromptPathShort, RemotePromptPathAbsolute),
		})
	}

	return issues
}

func validateAI(cfg *Config) []ValidationIssue {
	var issues []ValidationIssue

	if cfg.AI.Default != "" {
		if err := validateCloudDefault(cfg.AI.Default); err != nil {
			issues = append(issues, ValidationIssue{Field: "ai.default", Message: err.Error()})
		}
	}

	for _, id := range cfg.AI.Fallback {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if !IsCloudProvider(id) {
			issues = append(issues, ValidationIssue{
				Field:   "ai.fallback",
				Message: fmt.Sprintf("unknown provider %q in list", id),
			})
		}
	}

	for id, p := range cfg.AI.Providers {
		prefix := keyAIProvidersPrefix + id + "."
		if !IsKnownProvider(id) {
			issues = append(issues, ValidationIssue{
				Field:   keyAIProvidersPrefix + id,
				Message: "unknown provider",
			})
			continue
		}
		issues = append(issues, validateProvider(id, p, prefix)...)
	}

	issues = append(issues, validateTiered(cfg)...)
	issues = append(issues, validateAgent(cfg)...)
	return issues
}

func validateAgent(cfg *Config) []ValidationIssue {
	var issues []ValidationIssue

	if cfg.AI.Agent.MaxSteps < 0 {
		issues = append(issues, ValidationIssue{
			Field:   "ai.agent.max_steps",
			Message: "max_steps must be a non-negative integer",
		})
	}

	raw := strings.TrimSpace(cfg.AI.Agent.CommandTimeout)
	if raw != "" {
		if _, err := time.ParseDuration(raw); err != nil {
			issues = append(issues, ValidationIssue{
				Field:   keyAIAgentCommandTimeout,
				Message: fmt.Sprintf("invalid duration %q", cfg.AI.Agent.CommandTimeout),
			})
		}
	}

	return issues
}

func validateProvider(id string, p ProviderConfig, prefix string) []ValidationIssue {
	var issues []ValidationIssue

	style := strings.ToLower(strings.TrimSpace(p.APIStyle))
	if style == "" {
		style = APIStyleOpenAI
	} else if style != APIStyleOpenAI && style != APIStyleAnthropic {
		issues = append(issues, ValidationIssue{
			Field:   prefix + "api_style",
			Message: "api_style must be openai or anthropic",
		})
	} else if id == ProviderDeepseek && style == APIStyleAnthropic {
		issues = append(issues, ValidationIssue{
			Field:   prefix + "api_style",
			Message: "deepseek provider only supports openai api_style",
		})
	} else if id == ProviderOllama && style != APIStyleOpenAI {
		issues = append(issues, ValidationIssue{
			Field:   prefix + "api_style",
			Message: "ollama provider only supports openai api_style",
		})
	}

	baseURL := strings.TrimSpace(p.BaseURL)
	if baseURL == "" {
		baseURL = defaultBaseURL(id, style)
	}
	if err := validateProviderBaseURL(id, style, baseURL); err != nil {
		issues = append(issues, ValidationIssue{
			Field:   prefix + "base_url",
			Message: err.Error(),
		})
	}

	if strings.TrimSpace(p.Model) == "" {
		issues = append(issues, ValidationIssue{
			Field:   prefix + "model",
			Message: "model must not be empty",
		})
	}

	if p.ModelsDir != "" {
		if id != ProviderOllama {
			issues = append(issues, ValidationIssue{
				Field:   prefix + "models_dir",
				Message: "models_dir is only valid for provider ollama",
			})
		} else if _, err := expandAbsPath(p.ModelsDir); err != nil {
			issues = append(issues, ValidationIssue{
				Field:   prefix + "models_dir",
				Message: err.Error(),
			})
		}
	}

	return issues
}

func validateTiered(cfg *Config) []ValidationIssue {
	t := cfg.AI.Tiered
	var issues []ValidationIssue

	if t.LocalProvider != "" && t.LocalProvider != ProviderOllama {
		issues = append(issues, ValidationIssue{
			Field:   "ai.tiered.local_provider",
			Message: fmt.Sprintf("local_provider must be %q", ProviderOllama),
		})
	}

	if t.MaxPromptChars < 0 {
		issues = append(issues, ValidationIssue{
			Field:   "ai.tiered.max_prompt_chars",
			Message: "max_prompt_chars must be a positive integer",
		})
	} else if t.Enabled && t.MaxPromptChars <= 0 {
		issues = append(issues, ValidationIssue{
			Field:   "ai.tiered.max_prompt_chars",
			Message: "max_prompt_chars must be a positive integer",
		})
	}

	if t.MaxHistoryTurns < 0 {
		issues = append(issues, ValidationIssue{
			Field:   "ai.tiered.max_history_turns",
			Message: "max_history_turns must be a non-negative integer",
		})
	}

	return issues
}

func legacyDeepseekWarning(cfg *Config) []ValidationIssue {
	if cfg.Deepseek.BaseURL == "" && cfg.Deepseek.Model == "" {
		return nil
	}
	return []ValidationIssue{{
		Warning: true,
		Message: "legacy top-level deepseek: merged on load; consider moving to ai.providers.deepseek",
	}}
}
