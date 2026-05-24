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

package errhint

import (
	"errors"
	"fmt"
	"strings"
)

// WithTry wraps cause with a blank line, explanation, and optional Try block.
// errors.Is(err, cause) remains true. Error strings do not include the AiT: prefix.
func WithTry(cause error, explanation string, tries ...string) error {
	if cause == nil {
		return nil
	}

	var extra strings.Builder
	if explanation != "" {
		extra.WriteString("\n\n")
		extra.WriteString(explanation)
	}
	if block := FormatTryBlock(tries...); block != "" {
		extra.WriteString("\n\n")
		extra.WriteString(block)
	}
	if extra.Len() == 0 {
		return cause
	}
	return fmt.Errorf("%w%s", cause, extra.String())
}

// WithTryMsg is like WithTry but uses errors.New(msg) as the cause.
func WithTryMsg(msg, explanation string, tries ...string) error {
	return WithTry(errors.New(msg), explanation, tries...)
}

// FormatTryBlock builds the Try section (header plus indented command lines).
// Returns an empty string when tries is empty.
func FormatTryBlock(tries ...string) string {
	if len(tries) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(tryHeader)
	for _, line := range tries {
		if line == "" {
			continue
		}
		b.WriteString("\n  ")
		b.WriteString(line)
	}
	return b.String()
}
