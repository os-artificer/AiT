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
	"net/url"
	"strconv"
	"strings"
)

var unsupportedFlags = map[string]string{
	"-L": "local port forwarding (-L)",
	"-R": "remote port forwarding (-R)",
	"-D": "dynamic port forwarding (-D)",
	"-X": "X11 forwarding (-X)",
	"-Y": "trusted X11 forwarding (-Y)",
	"-f": "background mode (-f)",
	"-N": "no remote command (-N)",
	"-M": "control master (-M)",
	"-S": "control path (-S)",
	"-O": "control command (-O)",
	"-W": "stdio forward (-W)",
	"-w": "tun forwarding (-w)",
}

var noopFlags = map[string]bool{
	"-4": true, "-6": true, "-t": true, "-T": true, "-C": true, "-v": true,
	"-A": true, "-a": true, "-q": true, "-Q": true,
}

const connectArgvLabel = "" // default connect route; not the removed login subcommand

// Spec holds parsed login arguments (OpenSSH + AiT extensions).
type Spec struct {
	Destination   string
	RemoteCommand []string

	Port       int
	User       string
	Identity   string
	ConfigFile string
	ProxyJump  []string
	Options    map[string]string

	Alias    string
	Save     bool
	Tmux     bool
	Password string

	// Legacy AiT long flags (--host etc.)
	LegacyHost     string
	LegacyUser     string
	LegacyPort     int
	LegacyIdentity string
}

// ParseLogin parses argv for the default aissh connect route (ssh-compatible flags + AiT extensions).
// Flags may appear before or after destination until the first remote command
// argument; after that, tokens are passed through to RemoteCommand unchanged.
// Returns fmt.Errorf for unknown/unsupported flags, missing flag values, or invalid combinations.
func ParseLogin(argv []string) (Spec, error) {
	var s Spec
	s.Options = make(map[string]string)
	s.LegacyPort = -1

	remoteCommandStarted := false
	for i := 0; i < len(argv); i++ {
		arg := argv[i]
		if arg == "--" {
			s.RemoteCommand = append(s.RemoteCommand, argv[i+1:]...)
			break
		}
		if strings.HasPrefix(arg, "-") {
			if remoteCommandStarted {
				s.RemoteCommand = append(s.RemoteCommand, arg)
				continue
			}
			next, err := applyFlag(&s, argv, i, connectArgvLabel)
			if err != nil {
				return Spec{}, err
			}
			i = next
			continue
		}

		if s.Destination != "" {
			remoteCommandStarted = true
			s.RemoteCommand = append(s.RemoteCommand, arg)
			continue
		}
		s.Destination = arg
	}

	if err := applyLegacy(&s); err != nil {
		return Spec{}, err
	}

	if s.Destination == "" {
		return Spec{}, destinationRequiredErr()
	}
	return s, nil
}

// ParseDestination splits dest into user, host, and port.
// Supports [user@]host, host:port, user@[ipv6]:port, and ssh:// URIs; port is 0 when omitted.
// Returns user, host, port, and nil on success. Errors: fmt.Errorf for invalid destination syntax or port.
func ParseDestination(dest string) (user, host string, port int, err error) {
	dest = strings.TrimSpace(dest)
	if strings.HasPrefix(dest, "ssh://") {
		u, perr := url.Parse(dest)
		if perr != nil {
			return "", "", 0, fmt.Errorf("invalid ssh URI: %w", perr)
		}
		host = u.Hostname()
		if u.User != nil {
			user = u.User.Username()
		}
		if p := u.Port(); p != "" {
			port, _ = strconv.Atoi(p)
		}
		return user, host, port, nil
	}

	// user@host or user@[ipv6]:port
	if at := strings.LastIndex(dest, "@"); at >= 0 {
		user = dest[:at]
		dest = dest[at+1:]
	}

	hostPart := dest
	if strings.HasPrefix(dest, "[") {
		hostPart, port, err = parseBracketHost(dest)
		if err != nil {
			return "", "", 0, err
		}
	} else if colon := strings.LastIndex(dest, ":"); colon > 0 {
		if p, perr := strconv.Atoi(dest[colon+1:]); perr == nil {
			hostPart = dest[:colon]
			port = p
		}
	}
	host = hostPart
	return user, host, port, nil
}
