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
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"

	"gopkg.in/yaml.v3"
)

// ValidationIssue is one validation finding. Warning issues do not fail Validate OK status.
type ValidationIssue struct {
	Field   string
	Message string
	Warning bool
}

// ValidationReport holds the outcome of Validate for one config file path.
type ValidationReport struct {
	Path   string
	Issues []ValidationIssue
}

// OK reports whether there are no non-warning validation issues.
func (r *ValidationReport) OK() bool {
	if r == nil {
		return true
	}
	for _, iss := range r.Issues {
		if !iss.Warning {
			return false
		}
	}
	return true
}

// Error returns a multi-line summary of non-warning issues for tests.
func (r *ValidationReport) Error() string {
	if r == nil || r.OK() {
		return ""
	}

	var b strings.Builder
	for _, iss := range r.Issues {
		if iss.Warning {
			continue
		}
		if iss.Field != "" {
			fmt.Fprintf(&b, "%s: %s\n", iss.Field, iss.Message)
		} else {
			fmt.Fprintf(&b, "%s\n", iss.Message)
		}
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// Validate reads config.yaml at path and checks YAML syntax and semantic rules.
// When path is empty, platform.ConfigPath is used. Missing files return an errhint error (no auto-init).
func Validate(path string) (*ValidationReport, error) {
	if path == "" {
		var err error
		path, err = platform.ConfigPath()
		if err != nil {
			return nil, err
		}
	}

	report := &ValidationReport{Path: path}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return nil, configNotFoundErr(path)
	} else if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		report.Issues = append(report.Issues, ValidationIssue{
			Field:   "yaml",
			Message: err.Error(),
		})
		return report, nil
	}

	report.Issues = append(report.Issues, validateSemantics(&cfg)...)
	return report, nil
}

// FormatValidation writes a human-readable OK line or numbered issue list to w.
func FormatValidation(w io.Writer, report *ValidationReport) {
	if report == nil {
		return
	}
	if report.OK() {
		_, _ = fmt.Fprintf(w, "config OK: %s\n", report.Path)
		return
	}

	n := countErrors(report)
	word := "error"
	if n != 1 {
		word = "errors"
	}
	_, _ = fmt.Fprintf(w, "validation failed (%d %s) in %s:\n", n, word, report.Path)
	for _, iss := range report.Issues {
		if iss.Warning {
			_, _ = fmt.Fprintf(w, "  warning: %s\n", iss.Message)
			continue
		}
		if iss.Field != "" {
			_, _ = fmt.Fprintf(w, "  %s: %s\n", iss.Field, iss.Message)
		} else {
			_, _ = fmt.Fprintf(w, "  %s\n", iss.Message)
		}
	}
}

func countErrors(report *ValidationReport) int {
	n := 0
	for _, iss := range report.Issues {
		if !iss.Warning {
			n++
		}
	}
	return n
}
