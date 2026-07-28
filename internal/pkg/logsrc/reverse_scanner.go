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
	"io"
	"os"
	"strings"
)

// reverseChunkSize is how much of a file is read per backward step.
const reverseChunkSize = 64 * 1024

// reverseScanner yields the lines of one file last-line-first. Seekable files
// are read in reverseChunkSize blocks from the end, so `-n 20` on a
// multi-gigabyte log touches only the tail. Gzip streams cannot be seeked, so
// they are decoded once into memory.
type reverseScanner struct {
	file    *os.File
	off     int64    // start of the region not yet read
	pending []byte   // partial first line of the region already read
	queue   []string // lines ready to hand out, newest first
	qi      int
	atStart bool
}

// newReverseScanner opens path for backward reading. Returns the underlying
// open/gzip error, which callers map to ErrPermission where relevant.
func newReverseScanner(path string) (*reverseScanner, error) {
	if strings.HasSuffix(path, gzipSuffix) {
		return newGzipReverseScanner(path)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &reverseScanner{file: f, off: info.Size()}, nil
}

// newGzipReverseScanner decodes a gzip log fully, since it cannot be seeked.
func newGzipReverseScanner(path string) (*reverseScanner, error) {
	rc, err := openLogReader(path)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return &reverseScanner{queue: splitLinesReverse(data), atStart: true}, nil
}

// next returns the next line towards the start of the file. The second result
// is false at the start of the file.
func (s *reverseScanner) next() (string, bool, error) {
	for {
		if s.qi < len(s.queue) {
			line := s.queue[s.qi]
			s.qi++
			return line, true, nil
		}
		if s.atStart || s.file == nil {
			return "", false, nil
		}
		if err := s.fill(); err != nil {
			return "", false, err
		}
	}
}

// close releases the file, if any.
func (s *reverseScanner) close() error {
	if s.file == nil {
		return nil
	}
	f := s.file
	s.file = nil
	return f.Close()
}

// fill reads the block preceding s.off and queues its complete lines
// newest-first, carrying any partial leading line over to the next call.
func (s *reverseScanner) fill() error {
	size := int64(reverseChunkSize)
	if s.off < size {
		size = s.off
	}
	start := s.off - size

	buf := make([]byte, size)
	if _, err := s.file.ReadAt(buf, start); err != nil && err != io.EOF {
		return err
	}
	s.off = start

	parts := bytes.Split(append(buf, s.pending...), []byte("\n"))
	s.queue = s.queue[:0]
	s.qi = 0

	// parts[0] is only a complete line once the file start is reached.
	for i := len(parts) - 1; i >= 1; i-- {
		if line := trimLine(parts[i]); line != "" {
			s.queue = append(s.queue, line)
		}
	}
	if start > 0 {
		s.pending = append([]byte(nil), parts[0]...)
		return nil
	}

	s.atStart = true
	s.pending = nil
	if line := trimLine(parts[0]); line != "" {
		s.queue = append(s.queue, line)
	}
	return nil
}
