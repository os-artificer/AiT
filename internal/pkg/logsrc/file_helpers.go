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
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// isoLayouts are the timestamp layouts tried for ISO-style log prefixes.
var isoLayouts = []string{time.RFC3339Nano, time.RFC3339}

const (
	// gzipSuffix marks a compressed rotation file.
	gzipSuffix = ".gz"
	// scanLineMax bounds one log line held in memory.
	scanLineMax = 10 * 1024 * 1024
	// rfc3164StampLen is the width of a "Jan _2 15:04:05" timestamp.
	rfc3164StampLen = 15
	// futureSkew is how far ahead a stamp may be before the year-less
	// RFC3164 timestamp is treated as belonging to the previous year.
	futureSkew = 25 * time.Hour
	// rotationMaxIndex orders rotation suffixes (.1, .2) before the live file.
	rotationMaxIndex = 10000
	// scanInitialBuf is the starting line buffer, grown up to scanLineMax.
	scanInitialBuf = 64 * 1024
	// rfc3164Layout is the year-less timestamp classic syslog writes.
	rfc3164Layout = "Jan _2 15:04:05"
	// fileSourceLabel marks records that came from a plain-text log file.
	fileSourceLabel = "file"
	// globMeta are the characters that make a path a glob pattern.
	globMeta = "*?["
	// liveFileOrderKey sorts the un-rotated file after its rotations.
	liveFileOrderKey = "/zzzzz"
)

// ResolveLogFiles expands globs, adds sibling rotations of non-glob paths, and
// returns absolute paths ordered oldest → newest. Returns a filepath.Glob
// error for a malformed pattern.
func ResolveLogFiles(patterns []string) ([]string, error) {
	var out []string
	seen := map[string]struct{}{}

	for _, p := range patterns {
		matches, err := expandPattern(p)
		if err != nil {
			return nil, err
		}
		out = append(out, dedupPaths(matches, seen)...)

		if !strings.ContainsAny(p, globMeta) {
			out = append(out, findRotations(p, seen)...)
		}
	}

	sort.Slice(out, func(i, j int) bool {
		return rotationOrderKey(out[i]) < rotationOrderKey(out[j])
	})
	return out, nil
}

// expandPattern resolves one -S pattern to existing paths. A pattern with no
// glob matches is kept when it names an existing file, so that a plain path is
// still opened. Returns a filepath.Glob error for a malformed pattern.
func expandPattern(pattern string) ([]string, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(matches) > 0 {
		return matches, nil
	}
	if _, err := os.Stat(pattern); err != nil {
		return nil, nil
	}
	return []string{pattern}, nil
}

// dedupPaths absolutizes paths and drops ones already in seen.
func dedupPaths(paths []string, seen map[string]struct{}) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if abs, ok := addPath(p, seen); ok {
			out = append(out, abs)
		}
	}
	return out
}

// ParseSyslogLine parses a classic syslog / macOS system.log line. The second
// result is false only for blank lines: an unrecognized line is still returned
// as a record with the whole line as Message, a zero Time and LevelUnset, so
// that filters treat its unknown attributes as "not excluded".
func ParseSyslogLine(line string) (Record, bool) {
	line = strings.TrimRight(line, "\r\n")
	if line == "" {
		return Record{}, false
	}
	rec := Record{Raw: line, Message: line, Level: LevelUnset, Source: "file"}

	if parsed, ok := parseRFC3164(line); ok {
		return parsed, true
	}
	if fields := strings.SplitN(line, " ", 2); len(fields) == 2 {
		for _, layout := range isoLayouts {
			t, err := time.Parse(layout, fields[0])
			if err != nil {
				continue
			}
			rec.Time = t
			rec.Message = fields[1]
			return rec, true
		}
	}
	return rec, true
}

// parseRFC3164 parses "Jan  2 15:04:05 host tag[pid]: message". The format
// carries no year, so the current year is assumed and rolled back when that
// would place the entry in the future.
func parseRFC3164(line string) (Record, bool) {
	if len(line) <= rfc3164StampLen {
		return Record{}, false
	}
	stamp, err := time.ParseInLocation("Jan _2 15:04:05", line[:rfc3164StampLen], time.Local)
	if err != nil {
		return Record{}, false
	}

	now := time.Now()
	ts := time.Date(now.Year(), stamp.Month(), stamp.Day(),
		stamp.Hour(), stamp.Minute(), stamp.Second(), 0, time.Local)
	if ts.Sub(now) > futureSkew {
		ts = ts.AddDate(-1, 0, 0)
	}

	rec := Record{Raw: line, Message: line, Level: LevelUnset, Source: "file", Time: ts}
	fields := strings.SplitN(strings.TrimSpace(line[rfc3164StampLen:]), " ", 3)
	if len(fields) >= 1 {
		rec.Host = fields[0]
	}
	if len(fields) >= 2 {
		rec.Unit, rec.PID = parseSyslogTag(fields[1])
		rec.Process = rec.Unit
	}
	if len(fields) >= 3 {
		rec.Message = fields[2]
	}
	return rec, true
}

// parseSyslogTag splits "sshd[1234]:" into unit and pid.
func parseSyslogTag(tag string) (string, int) {
	i := strings.IndexByte(tag, '[')
	if i < 0 {
		return strings.TrimSuffix(tag, ":"), 0
	}
	pidStr := strings.TrimSuffix(strings.TrimSuffix(tag[i+1:], "]:"), "]")
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return tag[:i], 0
	}
	return tag[:i], pid
}

// openLogReader opens a log file, transparently decompressing .gz.
func openLogReader(path string) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(path, gzipSuffix) {
		return f, nil
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	return &gzipReadCloser{gz: gz, file: f}, nil
}

// newLogScanner returns a line scanner sized for long log lines.
func newLogScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), scanLineMax)
	return sc
}

// splitLinesReverse returns the non-empty lines of data, last line first.
func splitLinesReverse(data []byte) []string {
	parts := bytes.Split(data, []byte("\n"))
	out := make([]string, 0, len(parts))
	for i := len(parts) - 1; i >= 0; i-- {
		if line := trimLine(parts[i]); line != "" {
			out = append(out, line)
		}
	}
	return out
}

func trimLine(b []byte) string {
	return strings.TrimRight(string(b), "\r")
}

func addPath(path string, seen map[string]struct{}) (string, bool) {
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	if _, ok := seen[abs]; ok {
		return "", false
	}
	seen[abs] = struct{}{}
	return abs, true
}

// findRotations collects "<base>.N" and "<base>.N.gz" siblings of base.
func findRotations(base string, seen map[string]struct{}) []string {
	dir := filepath.Dir(base)
	name := filepath.Base(base)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var extra []string
	for _, e := range entries {
		n := e.Name()
		if n == name || !strings.HasPrefix(n, name+".") {
			continue
		}
		if abs, ok := addPath(filepath.Join(dir, n), seen); ok {
			extra = append(extra, abs)
		}
	}
	return extra
}

// rotationOrderKey sorts higher rotation numbers (older) before the live file.
func rotationOrderKey(path string) string {
	base := strings.TrimSuffix(filepath.Base(path), gzipSuffix)
	i := strings.LastIndex(base, ".")
	if i < 0 {
		return path + "/zzzzz"
	}
	n, err := strconv.Atoi(base[i+1:])
	if err != nil {
		return path + "/zzzzz"
	}
	return fmt.Sprintf("%s/%05d", filepath.Dir(path), rotationMaxIndex-n)
}
