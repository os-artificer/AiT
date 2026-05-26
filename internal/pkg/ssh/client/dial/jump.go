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

package dial

import (
	"fmt"
	"strings"
)

// ParseJumpDestination splits a ProxyJump destination (user@host:port) into its parts.
// Empty user or zero port mean "unset" (the caller substitutes defaults).
// Returns an error only when an IPv6 destination is malformed.
func ParseJumpDestination(dest string) (user, host string, port int, err error) {
	dest = strings.TrimSpace(dest)
	if at := strings.LastIndex(dest, "@"); at >= 0 {
		user = dest[:at]
		dest = dest[at+1:]
	}

	if strings.HasPrefix(dest, "[") {
		end := strings.Index(dest, "]")
		if end < 0 {
			return "", "", 0, fmt.Errorf("invalid jump host %q", dest)
		}
		host = dest[1:end]
		rest := dest[end+1:]
		if strings.HasPrefix(rest, ":") {
			fmt.Sscanf(rest[1:], "%d", &port)
		}
		return user, host, port, nil
	}

	if i := strings.LastIndex(dest, ":"); i > 0 {
		if p, perr := ParsePortString(dest[i+1:]); perr == nil {
			host = dest[:i]
			port = p
			return user, host, port, nil
		}
	}
	host = dest
	return user, host, 0, nil
}

// ParsePortString parses s as a decimal port number. Returns an Sscanf error
// when s does not start with a number.
func ParsePortString(s string) (int, error) {
	var p int
	_, err := fmt.Sscanf(s, "%d", &p)
	return p, err
}
