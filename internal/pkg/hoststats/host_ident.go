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

package hoststats

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/shirou/gopsutil/v4/host"
)

const (
	headerUserLabel   = "USER:"
	headerOSLabel     = "OS:"
	headerCPULabel    = "CPU:"
	headerMEMLabel    = "MEM:"
	headerKernelLabel = "KERNEL:"
	bytesPerMiB       = 1024 * 1024

	// HeaderFieldSep joins header groups (USER, OS, …, clock).
	HeaderFieldSep = "  "
	// HeaderClockLayout is the dashboard datetime format (local wall time).
	HeaderClockLayout = "2006/01/02 15:04:05"
)

// CollectHostIdent reads hostname, logged-in account, OS/arch, kernel, and CPU count.
// Individual fields may be empty when the platform denies that lookup.
func CollectHostIdent() HostIdent {
	ident := HostIdent{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
		CPUs: runtime.NumCPU(),
	}
	if name, err := os.Hostname(); err == nil {
		ident.Hostname = name
	}
	if u, err := user.Current(); err == nil && u != nil {
		ident.User = u.Username
		ident.UID = u.Uid
	}
	applyHostInfo(&ident)
	return ident
}

// FormatHostHeader builds a single dashboard/snapshot header line from ident, memTotal, and t.
// Fields are USER / OS / CPU / MEM / KERNEL plus datetime; hostname is omitted.
// Empty fields are skipped. memTotal is shown as integer MiB with a spaced MB unit (e.g. "1024 MB").
// t is formatted as local YYYY/MM/DD HH:MM:SS when non-zero.
func FormatHostHeader(ident HostIdent, memTotal uint64, t time.Time) string {
	return strings.Join(HostHeaderGroups(ident, memTotal, t), HeaderFieldSep)
}

// HostHeaderGroups returns labeled header tokens (and datetime) for wrapping and display.
// Empty fields are omitted. Hostname is never included.
func HostHeaderGroups(ident HostIdent, memTotal uint64, t time.Time) []string {
	parts := make([]string, 0, 6)
	if ident.User != "" {
		parts = append(parts, headerUserLabel+" "+ident.User)
	}
	if osArch := formatOSArch(ident.OS, ident.Arch); osArch != "" {
		parts = append(parts, headerOSLabel+" "+osArch)
	}
	if ident.CPUs > 0 {
		parts = append(parts, fmt.Sprintf("%s %d", headerCPULabel, ident.CPUs))
	}
	if memTotal > 0 {
		parts = append(parts, fmt.Sprintf("%s %d MB", headerMEMLabel, memTotal/bytesPerMiB))
	}
	if ident.Kernel != "" {
		parts = append(parts, headerKernelLabel+" "+ident.Kernel)
	}
	if !t.IsZero() {
		parts = append(parts, t.Format(HeaderClockLayout))
	}
	return parts
}

// WrapHostHeader packs groups into left-aligned lines that fit width (rune cells).
// Groups are never split across lines. width < 1 means unlimited (single line).
// Returns at least one empty string when groups is empty.
func WrapHostHeader(groups []string, width int) []string {
	if len(groups) == 0 {
		return []string{""}
	}
	if width < 1 {
		return []string{strings.Join(groups, HeaderFieldSep)}
	}

	var lines []string
	var cur []string
	curLen := 0
	sepLen := utf8.RuneCountInString(HeaderFieldSep)

	for _, g := range groups {
		gLen := utf8.RuneCountInString(g)
		need := gLen
		if len(cur) > 0 {
			need += sepLen
		}
		if len(cur) > 0 && curLen+need > width {
			lines = append(lines, strings.Join(cur, HeaderFieldSep))
			cur = []string{g}
			curLen = gLen
			continue
		}
		cur = append(cur, g)
		curLen += need
	}
	if len(cur) > 0 {
		lines = append(lines, strings.Join(cur, HeaderFieldSep))
	}
	return lines
}

func applyHostInfo(ident *HostIdent) {
	info, err := host.Info()
	if err != nil || info == nil {
		return
	}
	if ident.Hostname == "" {
		ident.Hostname = info.Hostname
	}
	ident.Platform = info.Platform
	ident.PlatformVersion = info.PlatformVersion
	ident.Kernel = info.KernelVersion
}

func formatOSArch(osName, arch string) string {
	switch {
	case osName != "" && arch != "":
		return osName + "/" + arch
	case osName != "":
		return osName
	default:
		return arch
	}
}
