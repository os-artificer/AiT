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

package history

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const historyFileMode = 0o600

// LoadFromFile replaces in-memory lines from path. Missing files are ignored.
func (h *LineHistory) LoadFromFile(path string) error {
	if h == nil {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			h.flushed = 0
			return nil
		}
		return err
	}
	defer f.Close()

	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	if len(lines) > h.cap {
		lines = lines[len(lines)-h.cap:]
	}
	h.lines = lines
	h.flushed = len(lines)
	h.filePath = path
	return nil
}

// AppendPendingToFile appends in-memory lines not yet flushed (history -a).
func (h *LineHistory) AppendPendingToFile() error {
	if h == nil || h.filePath == "" {
		return nil
	}
	for i := h.flushed; i < len(h.lines); i++ {
		if err := appendLineToFile(h.filePath, h.lines[i]); err != nil {
			return err
		}
	}
	h.flushed = len(h.lines)
	return nil
}

// WriteFile overwrites the persistence file with all in-memory lines (history -w).
func (h *LineHistory) WriteFile() error {
	if h == nil || h.filePath == "" {
		return nil
	}
	if err := writeLinesToFile(h.filePath, h.lines); err != nil {
		return err
	}
	h.flushed = len(h.lines)
	return nil
}

// ReloadFromFile discards memory and reloads from the persistence file (history -r).
func (h *LineHistory) ReloadFromFile() error {
	if h == nil || h.filePath == "" {
		return fmt.Errorf("history: no history file")
	}
	h.lines = nil
	h.flushed = 0
	return h.LoadFromFile(h.filePath)
}

func appendLineToFile(path, line string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, historyFileMode)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintln(f, line)
	return err
}

func writeLinesToFile(path string, lines []string) error {
	var b strings.Builder
	for _, line := range lines {
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(b.String()), historyFileMode)
}
