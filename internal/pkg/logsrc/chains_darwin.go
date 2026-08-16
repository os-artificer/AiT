//go:build darwin

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

func platformChains() map[string][]Layer {
	return map[string][]Layer{
		SourceNameSystem: {
			nativeLogShow(true),
			nativeLogShow(false),
			builtinFiles("system.log", []string{"/var/log/system.log"}),
			deferredLayer(ProviderASL, ReasonNotImplemented),
		},
		SourceNameInstall: {
			builtinFiles("install.log", []string{"/var/log/install.log"}),
		},
	}
}

// nativeLogShow builds a `log show` layer. ndjson selects the structured
// output, which parses reliably; the syslog style behind it is the fallback for
// older macOS releases that do not support --style ndjson.
func nativeLogShow(ndjson bool) Layer {
	style := LogShowStyleNDJSON
	if !ndjson {
		style = LogShowStyleSyslog
	}
	return Layer{
		Kind:      LayerNative,
		Provider:  ToolLog + " show --style " + style,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			if _, err := r.lookPath(ToolLog); err != nil {
				return nil, fmtErrTool(ToolLog)
			}
			return &LogShowSource{Resolver: r, NDJSON: ndjson}, nil
		},
	}
}
