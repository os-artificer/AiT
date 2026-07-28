//go:build windows

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

// Windows Event Log channel names behind the logical source names.
const (
	channelSystem      = "System"
	channelApplication = "Application"
	channelSecurity    = "Security"

	// providerChannelSep joins a backend name and a channel in provider labels.
	providerChannelSep = ":"

	// setupLogsProvider labels the text logs under C:\Windows\Logs.
	setupLogsProvider = "Windows Logs"
)

func platformChains() map[string][]Layer {
	return map[string][]Layer{
		SourceNameSystem:      eventChannelChain(channelSystem),
		SourceNameApplication: eventChannelChain(channelApplication),
		SourceNameSecurity:    eventChannelChain(channelSecurity),
		SourceNameSetup: {
			builtinFiles(setupLogsProvider, []string{
				`C:\Windows\Logs\CBS\CBS.log`,
				`C:\Windows\Logs\*.log`,
			}),
		},
	}
}

// eventChannelChain prefers the in-process wevtapi.dll reader, falls back to
// the wevtutil binary, and names the .evtx reader that is not built yet.
func eventChannelChain(channel string) []Layer {
	return []Layer{
		builtinEvtAPI(channel),
		nativeWevtutil(channel),
		deferredLayer(ProviderEvtx, ReasonNotImplemented),
	}
}

// builtinEvtAPI builds the in-process wevtapi.dll layer, which needs no
// external executable.
func builtinEvtAPI(channel string) Layer {
	return Layer{
		Kind:      LayerBuiltin,
		Provider:  SourceWevtAPI + providerChannelSep + channel,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			return &EvtAPISource{Channel: channel, Label: channel}, nil
		},
	}
}

// nativeWevtutil builds the wevtutil.exe layer, used when the API is
// unavailable.
func nativeWevtutil(channel string) Layer {
	return Layer{
		Kind:      LayerNative,
		Provider:  ToolWevtutil + providerChannelSep + channel,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			if _, err := r.lookPath(ToolWevtutil); err != nil {
				return nil, fmtErrTool(ToolWevtutil)
			}
			return &WevtutilSource{Resolver: r, Channel: channel}, nil
		},
	}
}
