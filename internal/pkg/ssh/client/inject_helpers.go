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

package client

import "strings"

const (
	// injectSpacePrefix is prepended to all AiT-injected PTY commands.
	// With HISTCONTROL=ignorespace (bash) or HIST_IGNORE_SPACE (zsh),
	// space-prefixed commands are excluded from shell history.
	injectSpacePrefix = " "

	// injectEchoOff is the stty command to disable PTY echo before an injection.
	injectEchoOff = "stty -echo 2>/dev/null"

	// injectEchoOn is the stty command to re-enable PTY echo after an injection.
	injectEchoOn = "stty echo 2>/dev/null"

	// injectCaptureBegin is the plain-text marker injected before a captured command.
	// Uses plain text (not ANSI escape) so intermediate PTY drivers in chain mode
	// cannot strip or reinterpret the marker.
	injectCaptureBegin = "AIT_CAPTURE_B_f3e7a1c9"

	// injectCaptureEnd is the plain-text marker injected after a captured command.
	injectCaptureEnd = "AIT_CAPTURE_E_f3e7a1c9"

	// injectCaptureBeginPrint is the printf command to emit the capture-begin marker.
	injectCaptureBeginPrint = "printf 'AIT_CAPTURE_B_f3e7a1c9'"

	// injectCaptureEndPrint is the printf command to emit the capture-end marker.
	injectCaptureEndPrint = "printf 'AIT_CAPTURE_E_f3e7a1c9'"
)

// buildInjectParts constructs the shell command parts for a silent injection.
// opts controls echo suppression and output capture wrapping.
func buildInjectParts(cmd string, opts InjectOpts) []string {
	var parts []string

	if opts.SuppressEcho {
		parts = append(parts, injectSpacePrefix+injectEchoOff)
	}

	if opts.CaptureOutput {
		parts = append(parts, injectSpacePrefix+injectCaptureBeginPrint)
	}

	parts = append(parts, injectSpacePrefix+cmd)

	if opts.CaptureOutput {
		parts = append(parts, injectSpacePrefix+injectCaptureEndPrint)
	}

	if opts.SuppressEcho {
		parts = append(parts, injectSpacePrefix+injectEchoOn)
	}

	return parts
}

// injectStripFragments returns scrollback/filter fragments for injected command
// leakage suppression. These substrings are stripped from scrollback during
// capture so the user never sees the injected delimiters or echo commands.
func injectStripFragments() []string {
	return []string{
		injectCaptureBegin,
		injectCaptureEnd,
		injectCaptureBeginPrint,
		injectCaptureEndPrint,
	}
}

// parseCapturedOutput extracts content between the capture-begin and capture-end
// markers from the delta output. Uses LastIndex for the begin marker because the
// injected command line echo (which also contains the marker text) appears before
// the actual printf output in the PTY scrollback delta.
func parseCapturedOutput(delta string) string {
	begin := injectCaptureBegin
	end := injectCaptureEnd

	// The injected command line (e.g. "printf 'AIT_CAPTURE_B_f3e7a1c9'")
	// contains the marker text and is echoed to the PTY before the actual
	// printf output. Use LastIndex to find the real marker output.
	startIdx := strings.LastIndex(delta, begin)
	if startIdx < 0 {
		return delta
	}
	afterBegin := delta[startIdx+len(begin):]

	endIdx := strings.Index(afterBegin, end)
	if endIdx < 0 {
		return afterBegin
	}

	return afterBegin[:endIdx]
}

// parseHashFromOutput extracts a sha256sum hash from command output.
// The output format is: "hash  path" or "hash  -" (stdin).
// Returns the hash string, or empty string if parsing fails.
func parseHashFromOutput(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}

	// sha256sum output format: "hash  filename" — take the first field.
	fields := strings.Fields(output)
	if len(fields) < 1 {
		return ""
	}

	return fields[0]
}
