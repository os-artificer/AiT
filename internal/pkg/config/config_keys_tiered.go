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
)

func (c *Config) setTieredField(field, value string) error {
	normalizeTiered(c)
	switch field {
	case "enabled":
		v, err := parseBool(value)
		if err != nil {
			return err
		}
		c.AI.Tiered.Enabled = v
	case "local_provider":
		if value != ProviderOllama {
			return fmt.Errorf("local_provider must be %q", ProviderOllama)
		}
		c.AI.Tiered.LocalProvider = value
	case "max_prompt_chars":
		n, err := strconv.Atoi(value)
		if err != nil || n <= 0 {
			return errors.New("max_prompt_chars must be a positive integer")
		}
		c.AI.Tiered.MaxPromptChars = n
	case "max_history_turns":
		n, err := strconv.Atoi(value)
		if err != nil || n < 0 {
			return errors.New("max_history_turns must be a non-negative integer")
		}
		c.AI.Tiered.MaxHistoryTurns = n
	case "complex_keywords":
		c.AI.Tiered.ComplexKeywords = parseKeywordList(value)
	default:
		return errors.New("unknown tiered field: " + field)
	}
	return nil
}

func (c *Config) getTieredField(field string) (string, error) {
	normalizeTiered(c)
	t := c.AI.Tiered
	switch field {
	case "enabled":
		return strconv.FormatBool(t.Enabled), nil
	case "local_provider":
		return t.LocalProvider, nil
	case "max_prompt_chars":
		return strconv.Itoa(t.MaxPromptChars), nil
	case "max_history_turns":
		return strconv.Itoa(t.MaxHistoryTurns), nil
	case "complex_keywords":
		return strings.Join(t.ComplexKeywords, ","), nil
	default:
		return "", errors.New("unknown tiered field: " + field)
	}
}
