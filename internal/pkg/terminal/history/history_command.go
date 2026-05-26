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
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RunHistoryCommand executes bash-style history subcommands on h and writes listing to w.
func RunHistoryCommand(w io.Writer, h *LineHistory, args []string) error {
	if h == nil {
		return fmt.Errorf("history: not available")
	}
	if len(args) == 0 {
		return h.Format(w, 0)
	}

	switch args[0] {
	case "-c":
		if len(args) != 1 {
			return fmt.Errorf("history: -c takes no arguments")
		}
		h.Clear()
		return nil
	case "-d":
		if len(args) != 2 {
			return fmt.Errorf("history: usage: history -d offset")
		}
		n, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("history: invalid offset %q", args[1])
		}
		return h.Delete(n)
	case "-s":
		if len(args) < 2 {
			return fmt.Errorf("history: usage: history -s cmdline")
		}
		h.SetLine(strings.Join(args[1:], " "))
		return nil
	case "-a":
		if len(args) != 1 {
			return fmt.Errorf("history: -a takes no arguments")
		}
		return h.AppendPendingToFile()
	case "-r":
		if len(args) != 1 {
			return fmt.Errorf("history: -r takes no arguments")
		}
		return h.ReloadFromFile()
	case "-w":
		if len(args) != 1 {
			return fmt.Errorf("history: -w takes no arguments")
		}
		return h.WriteFile()
	default:
		if n, err := strconv.Atoi(args[0]); err == nil && n > 0 {
			return h.Format(w, n)
		}
		return fmt.Errorf("history: unknown option %q", args[0])
	}
}
