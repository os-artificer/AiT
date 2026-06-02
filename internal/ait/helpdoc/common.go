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

// Package helpdoc provides login-style CLI help text for ait commands.
package helpdoc

import (
	"fmt"
	"io"
	"strings"
)

const (
	longHint              = "Run 'ait %s --help' for usage and examples."
	helpUsageLineFmt      = "Usage:\n  %s\n\n"
	helpExamplesHeader    = "Examples:"
	helpFooterFlagsPrefix = "\nFlags:\n  -h, --help      help for %s\n"
)

type helpBody struct {
	w io.Writer
}

func newHelpBody(w io.Writer) *helpBody {
	return &helpBody{w: w}
}

// WriteIntro writes introductory lines followed by a blank line to the help body.
func (b *helpBody) WriteIntro(lines ...string) error {
	for _, line := range lines {
		if _, err := fmt.Fprintf(b.w, "%s\n", line); err != nil {
			return err
		}
	}
	if len(lines) > 0 {
		_, err := fmt.Fprintln(b.w)
		return err
	}
	return nil
}

// WriteUsage writes the Usage section with the given command-line example.
func (b *helpBody) WriteUsage(line string) error {
	if _, err := fmt.Fprintf(b.w, helpUsageLineFmt, line); err != nil {
		return err
	}
	return nil
}

// WriteSection writes a titled section; skips when body is blank.
func (b *helpBody) WriteSection(title string, body string) error {
	if strings.TrimSpace(body) == "" {
		return nil
	}
	if _, err := fmt.Fprintf(b.w, "%s:\n%s", title, body); err != nil {
		return err
	}
	if !strings.HasSuffix(body, "\n") {
		_, err := fmt.Fprintln(b.w)
		return err
	}
	_, err := fmt.Fprintln(b.w)
	return err
}

// WriteExamples writes the Examples section with indented example lines.
func (b *helpBody) WriteExamples(lines ...string) error {
	if len(lines) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(b.w, helpExamplesHeader); err != nil {
		return err
	}
	for _, line := range lines {
		if _, err := fmt.Fprintf(b.w, "  %s\n", line); err != nil {
			return err
		}
	}
	_, err := fmt.Fprintln(b.w)
	return err
}

// Finish writes the trailing Flags section with the -h/--help line for cmdName.
func (b *helpBody) Finish(cmdName string) error {
	_, err := io.WriteString(b.w, HelpFooter(cmdName))
	return err
}
