//go:build linux

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
			nativeJournalctl(false),
			deferredLayer(ProviderJournalFile, ReasonNotImplemented),
			builtinFiles("syslog", []string{"/var/log/syslog", "/var/log/messages"}),
			builtinKmsg(),
		},
		SourceNameKernel: {
			nativeJournalctl(true),
			builtinKmsg(),
			builtinFiles("kern.log", []string{"/var/log/kern.log"}),
		},
		SourceNameAuth: {
			nativeJournalctl(false),
			builtinFiles("auth", []string{"/var/log/auth.log", "/var/log/secure"}),
		},
	}
}

// nativeJournalctl builds the journalctl layer. kernel restricts the query to
// kernel messages (journalctl -k).
func nativeJournalctl(kernel bool) Layer {
	return Layer{
		Kind:      LayerNative,
		Provider:  ToolJournalctl,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			if _, err := r.lookPath(ToolJournalctl); err != nil {
				return nil, fmtErrTool(ToolJournalctl)
			}
			return &JournalctlSource{Resolver: r, Kernel: kernel}, nil
		},
	}
}

// builtinKmsg builds the kernel ring buffer layer, the last resort on Linux.
func builtinKmsg() Layer {
	return Layer{
		Kind:      LayerBuiltin,
		Provider:  SourceKmsg,
		Available: true,
		Open: func(r *Resolver, q Query) (Source, error) {
			return &KmsgSource{}, nil
		},
	}
}
