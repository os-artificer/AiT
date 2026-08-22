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

// LayerKind classifies a resolution layer.
type LayerKind string

const (
	// LayerNative runs an operating-system log tool.
	LayerNative LayerKind = "native"
	// LayerBuiltin uses a pure-Go reader with no external dependency.
	LayerBuiltin LayerKind = "builtin"
	// LayerDeferred is a placeholder for a binary store reader that is not
	// implemented yet; it is always reported unavailable with a reason.
	LayerDeferred LayerKind = "deferred"
)

// Logical source names shared by the platform chains and the CLI help.
const (
	// SourceNameSystem is the default logical source on every platform.
	SourceNameSystem = "system"
	// SourceNameKernel is the kernel ring buffer / kernel log.
	SourceNameKernel = "kernel"
	// SourceNameAuth is the authentication / security log.
	SourceNameAuth = "auth"
	// SourceNameInstall is the macOS installer log.
	SourceNameInstall = "install"
	// SourceNameApplication is the Windows Application channel.
	SourceNameApplication = "application"
	// SourceNameSecurity is the Windows Security channel.
	SourceNameSecurity = "security"
	// SourceNameSetup is the Windows setup / servicing text logs.
	SourceNameSetup = "setup"
)

// ReasonNotImplemented marks a deferred layer whose reader is not built yet.
const ReasonNotImplemented = "not implemented"

// Backend names, used both as the executable to look up on PATH and as the
// provider label shown by `ailog sources`.
const (
	// ToolJournalctl reads the systemd journal on Linux.
	ToolJournalctl = "journalctl"
	// ToolLog is the macOS unified logging client, `log show`.
	ToolLog = "log"
	// ToolWevtutil is the Windows Event Log command-line client.
	ToolWevtutil = "wevtutil"

	// SourceKmsg is the Linux kernel ring buffer reader.
	SourceKmsg = "kmsg"
	// SourceWevtAPI is the in-process Windows Event Log reader (wevtapi.dll).
	SourceWevtAPI = "wevtapi"

	// ProviderJournalFile names the deferred .journal binary reader.
	ProviderJournalFile = "journal-file"
	// ProviderASL names the deferred macOS ASL binary reader.
	ProviderASL = "asl"
	// ProviderEvtx names the deferred .evtx binary reader.
	ProviderEvtx = "evtx"
)

// Layer is one resolution step in a source chain.
type Layer struct {
	Kind      LayerKind
	Provider  string
	Reason    string // why Available is false (deferred or unsupported)
	Available bool
	Open      func(r *Resolver, q Query) (Source, error)
}

// Capability is one Probe row, as rendered by `ailog sources`.
type Capability struct {
	Source    string `json:"source"`
	Backend   string `json:"backend"`
	Provider  string `json:"provider"`
	Available bool   `json:"available"`
	Selected  bool   `json:"selected,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

// Chains returns the platform-specific source name → layer chain map. Layers
// are ordered most preferred first.
func Chains() map[string][]Layer {
	return platformChains()
}

// DefaultSourceName returns the source used when -S is omitted.
func DefaultSourceName() string {
	return SourceNameSystem
}
