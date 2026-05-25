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
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

func (c *Config) setProviderField(id, field, value string) error {
	ensureProviders(c)
	p := c.AI.Providers[id]

	switch field {
	case "enabled":
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		p.Enabled = &v
	case "base_url":
		if err := validateProviderBaseURL(id, p.APIStyle, value); err != nil {
			return err
		}
		p.BaseURL = value
	case "model":
		p.Model = value
	case "api_style":
		style := strings.ToLower(strings.TrimSpace(value))
		if style != APIStyleOpenAI && style != APIStyleAnthropic {
			return fmt.Errorf("api_style must be openai or anthropic")
		}
		if id == ProviderDeepseek && style == APIStyleAnthropic {
			return errors.New("deepseek provider only supports openai api_style")
		}
		if id == ProviderOllama && style != APIStyleOpenAI {
			return errors.New("ollama provider only supports openai api_style")
		}
		p.APIStyle = style
		if p.BaseURL == "" {
			p.BaseURL = defaultBaseURL(id, style)
		}
	case "models_dir":
		if id != ProviderOllama {
			return errors.New("models_dir is only valid for provider ollama")
		}
		dir, err := expandAbsPath(value)
		if err != nil {
			return err
		}
		p.ModelsDir = dir
	default:
		return errors.New("unknown provider field: " + field)
	}

	c.AI.Providers[id] = p
	return nil
}

func (c *Config) getProviderField(id, field string) (string, error) {
	p, ok := c.Provider(id)
	if !ok {
		p = defaultProviderConfig(id)
	}

	switch field {
	case "enabled":
		return strconv.FormatBool(c.ProviderEnabled(id)), nil
	case "base_url":
		return p.BaseURL, nil
	case "model":
		return p.Model, nil
	case "api_style":
		if p.APIStyle == "" {
			return APIStyleOpenAI, nil
		}
		return p.APIStyle, nil
	case "models_dir":
		return p.ModelsDir, nil
	default:
		return "", errors.New("unknown provider field: " + field)
	}
}

func expandAbsPath(raw string) (string, error) {
	abs, err := platform.ExpandUserAbsPath(raw)
	if err != nil {
		if errors.Is(err, platform.ErrEmptyPath) {
			return "", errors.New("models_dir must not be empty")
		}
		return "", err
	}
	return abs, nil
}
