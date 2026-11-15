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

package net

import (
	"io"

	"github.com/os-artificer/AiT/internal/host/shared"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// SectionFlagsNet is the Flags help body for aihost net (traffic + connections).
const SectionFlagsNet = shared.SectionFlagsJSONWatch +
	"      --conn               list network connections instead of traffic\n" +
	"      --proto string       connection proto: tcp|udp|all (default tcp; with --conn)\n" +
	"      --state string       comma-separated states (with --conn; exclusive with --listen)\n" +
	"      --listen             TCP LISTEN sockets only (not UDP binds; with --conn)\n" +
	"      --port int           match local or remote port 1-65535 (with --conn)\n" +
	"      --pid int            filter by process id (with --conn)\n" +
	"      --limit int          max connection rows (default 100, max 5000; with --conn)\n"

// WriteHelp prints aihost net --help to w.
func WriteHelp(w io.Writer) error {
	body := pkghelpdoc.NewBody(w)
	if err := body.WriteIntro(
		"Show host network traffic, or list connections with --conn.",
		"Default mode prints aggregate interface rates, errors, and drops.",
		"With --conn: PROTO/STATE/LOCAL/REMOTE/PID/PROCESS/AGE/CREATED.",
		"AGE/CREATED prefer Linux /proc/<pid>/fd lifetime; on failure may use nf_conntrack delta-time (CREATED then approximate); else -.",
		"--port matches local or remote port (default --proto tcp; use udp|all for non-TCP).",
		"--listen is TCP LISTEN only (UDP binds use status NONE and are excluded).",
		"--proto filters sockets (tcp|udp|all); not the aithost proto counters command.",
		"--limit with --conn caps rows (default 100, max 5000); separate from proc/top --limit.",
	); err != nil {
		return err
	}
	if err := body.WriteUsage(toolname.HostCLI("net [flags]")); err != nil {
		return err
	}
	if err := body.WriteSection(pkghelpdoc.SectionFlags, SectionFlagsNet); err != nil {
		return err
	}
	if err := body.WriteExamples(
		toolname.HostCLI("net"),
		toolname.HostCLI("net --json"),
		toolname.HostCLI("net --watch --interval 2s"),
		toolname.HostCLI("net --conn"),
		toolname.HostCLI("net --conn --port 8080"),
		toolname.HostCLI("net --conn --listen"),
		toolname.HostCLI("net --conn --state ESTABLISHED --port 443"),
		toolname.HostCLI("net --conn --proto udp --json"),
	); err != nil {
		return err
	}
	return body.Finish(Name)
}
