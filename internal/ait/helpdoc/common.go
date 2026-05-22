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

const longHint = "Run 'ait %s --help' for usage and examples."

// LongHint returns the standard one-line Cobra Long field for cmdName.
func LongHint(cmdName string) string {
	return fmt.Sprintf(longHint, cmdName)
}

// HelpRequested reports whether argv is only -h or --help.
func HelpRequested(argv []string) bool {
	return len(argv) == 1 && (argv[0] == "-h" || argv[0] == "--help")
}

// HelpFooter returns the standard trailing Flags section for cmdName.
func HelpFooter(cmdName string) string {
	return fmt.Sprintf("\nFlags:\n  -h, --help      help for %s\n", cmdName)
}

type helpBody struct {
	w io.Writer
}

func newHelpBody(w io.Writer) *helpBody {
	return &helpBody{w: w}
}

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

func (b *helpBody) WriteUsage(line string) error {
	if _, err := fmt.Fprintf(b.w, "Usage:\n  %s\n\n", line); err != nil {
		return err
	}
	return nil
}

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

func (b *helpBody) WriteExamples(lines ...string) error {
	if len(lines) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(b.w, "Examples:"); err != nil {
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

func (b *helpBody) Finish(cmdName string) error {
	_, err := io.WriteString(b.w, HelpFooter(cmdName))
	return err
}
