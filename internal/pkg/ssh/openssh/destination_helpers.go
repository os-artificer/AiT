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

package openssh

import (
	"fmt"
	"strconv"
	"strings"
)

// parseBracketHost extracts host and optional port from [host]:port or [host].
func parseBracketHost(dest string) (host string, port int, err error) {
	end := strings.Index(dest, "]")
	if end < 0 {
		return "", 0, fmt.Errorf("invalid destination %q", dest)
	}
	host = dest[1:end]
	rest := dest[end+1:]

	if !strings.HasPrefix(rest, ":") {
		if rest != "" {
			return "", 0, fmt.Errorf("invalid destination %q", dest)
		}
		return host, 0, nil
	}

	port, err = strconv.Atoi(rest[1:])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in %q", dest)
	}
	return host, port, nil
}
