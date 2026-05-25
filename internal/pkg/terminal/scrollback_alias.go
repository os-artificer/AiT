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

package terminal

import (
	"github.com/os-artificer/AiT/internal/pkg/terminal/scrollback"
)

// Buffer is the scrollback ring buffer (alias of scrollback.Buffer) preserved
// for compatibility with existing importers (cli/session/ssh/client).
type Buffer = scrollback.Buffer

// LoadFromFile reads at most the scrollback cap from path for replay or AI context.
// Errors: os.Open or io.ReadAll failures.
func LoadFromFile(path string) (string, error) {
	return scrollback.LoadFromFile(path)
}

// Replay writes data to os.Stdout, appending a newline when non-empty and lacking one.
func Replay(data string) {
	scrollback.Replay(data)
}
