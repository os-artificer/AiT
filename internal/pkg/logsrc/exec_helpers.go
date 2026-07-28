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

package logsrc

import (
	"bytes"
	"fmt"
	"strings"
)

// permissionMarkers are stderr fragments that mean "denied" across the native
// log tools on Linux, macOS and Windows.
var permissionMarkers = []string{
	"permission",
	"access is denied",
	"operation not permitted",
	"not authorized",
}

// missingMarkers are stderr fragments that mean "tool or channel not found".
var missingMarkers = []string{
	"not found",
	"no such file",
	"is not recognized",
}

// FrameEventXML returns a LineFramer that accumulates `wevtutil qe /f:xml`
// output until a complete <Event>…</Event> document has been seen, ignoring
// the surrounding <Events> wrapper. wevtutil may pretty-print one event across
// many lines, so line-per-record framing would drop every event.
func FrameEventXML() LineFramer {
	var buf bytes.Buffer
	return func(line []byte) ([]byte, bool) {
		trimmed := bytes.TrimSpace(line)
		if len(trimmed) == 0 {
			return nil, false
		}
		if bytes.HasPrefix(trimmed, []byte("<?xml")) ||
			bytes.Equal(trimmed, []byte("<Events>")) || bytes.Equal(trimmed, []byte("</Events>")) {
			return nil, false
		}

		buf.Write(trimmed)
		if !bytes.Contains(trimmed, []byte("</Event>")) {
			return nil, false
		}
		doc := append([]byte(nil), buf.Bytes()...)
		buf.Reset()
		return doc, true
	}
}

// MapExecWaitErr maps a child process failure to ErrPermission,
// ErrToolMissing, or waitErr annotated with its stderr.
func MapExecWaitErr(stderr string, waitErr error) error {
	if waitErr == nil {
		return nil
	}
	lower := strings.ToLower(stderr + " " + waitErr.Error())
	if containsAny(lower, permissionMarkers) {
		return fmt.Errorf("%w: %v", ErrPermission, waitErr)
	}
	if containsAny(lower, missingMarkers) {
		return fmt.Errorf("%w: %v", ErrToolMissing, waitErr)
	}
	if stderr != "" {
		return fmt.Errorf("%w: %s", waitErr, stderr)
	}
	return waitErr
}

// frameSingleLine is the default framer: one non-blank line is one document.
func frameSingleLine(line []byte) ([]byte, bool) {
	if len(bytes.TrimSpace(line)) == 0 {
		return nil, false
	}
	return line, true
}

func mapStartErr(err error) error {
	if err == nil {
		return nil
	}
	if containsAny(strings.ToLower(err.Error()), missingMarkers) {
		return fmt.Errorf("%w: %v", ErrToolMissing, err)
	}
	return err
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}
