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

package sshargs

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

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
	Tmux     bool
	Password string

	// Legacy AiT long flags (--host etc.)
	LegacyHost     string
	LegacyUser     string
	LegacyPort     int
	LegacyIdentity string
}

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

// Parse parses argv for ait login (ssh-compatible flags + AiT extensions).
// Flags may appear before or after destination until the first remote command
// argument; after that, tokens are passed through to RemoteCommand unchanged.
func Parse(argv []string) (Spec, error) {
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
			next, err := applyFlag(&s, argv, i)
			if err != nil {
				return Spec{}, err
			}
			i = next
			continue
		}
		if s.Destination == "" {
			s.Destination = arg
			continue
		}
		remoteCommandStarted = true
		s.RemoteCommand = append(s.RemoteCommand, arg)
	}

	if err := s.applyLegacy(); err != nil {
		return Spec{}, err
	}
	if s.Destination == "" {
		return Spec{}, fmt.Errorf("destination required (e.g. user@host or ssh config Host alias)")
	}
	return s, nil
}

// applyFlag consumes one flag at argv[i] and returns the index of the next token.
func applyFlag(s *Spec, argv []string, i int) (int, error) {
	arg := argv[i]
	if msg, ok := unsupportedFlags[arg]; ok {
		return i, fmt.Errorf("%s is not supported by ait login; use the system ssh client", msg)
	}
	if noopFlags[arg] {
		return i, nil
	}
	if glued, err := applyGluedFlag(s, arg); glued {
		return i, err
	}
	switch arg {
	case "-p":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-p requires a value")
		}
		port, err := strconv.Atoi(argv[i])
		if err != nil {
			return i, fmt.Errorf("invalid -p: %w", err)
		}
		s.Port = port
		return i, nil
	case "-l":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-l requires a value")
		}
		s.User = argv[i]
		return i, nil
	case "-i":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-i requires a value")
		}
		s.Identity = argv[i]
		return i, nil
	case "-F":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-F requires a value")
		}
		s.ConfigFile = argv[i]
		return i, nil
	case "-J":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-J requires a value")
		}
		for _, j := range strings.Split(argv[i], ",") {
			j = strings.TrimSpace(j)
			if j != "" {
				s.ProxyJump = append(s.ProxyJump, j)
			}
		}
		return i, nil
	case "-o":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("-o requires a value")
		}
		if err := mergeOption(s.Options, argv[i]); err != nil {
			return i, err
		}
		return i, nil
	case "--alias":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--alias requires a value")
		}
		s.Alias = argv[i]
		return i, nil
	case "--tmux":
		s.Tmux = true
		return i, nil
	case "--password":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--password requires a value")
		}
		s.Password = argv[i]
		return i, nil
	case "--host":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--host requires a value")
		}
		s.LegacyHost = argv[i]
		return i, nil
	case "--user":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--user requires a value")
		}
		s.LegacyUser = argv[i]
		return i, nil
	case "--port":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--port requires a value")
		}
		port, err := strconv.Atoi(argv[i])
		if err != nil {
			return i, fmt.Errorf("invalid --port: %w", err)
		}
		s.LegacyPort = port
		return i, nil
	case "--identity":
		i++
		if i >= len(argv) {
			return i, fmt.Errorf("--identity requires a value")
		}
		s.LegacyIdentity = argv[i]
		return i, nil
	default:
		if strings.HasPrefix(arg, "-o") && len(arg) > 2 {
			if err := mergeOption(s.Options, arg[2:]); err != nil {
				return i, err
			}
			return i, nil
		}
		return i, fmt.Errorf("unknown flag %q", arg)
	}
}

// applyGluedFlag handles OpenSSH glued forms like -p60528 or -i~/.ssh/id_ed25519.
func applyGluedFlag(s *Spec, arg string) (glued bool, err error) {
	if len(arg) < 3 || arg[0] != '-' {
		return false, nil
	}
	suffix := arg[2:]
	switch arg[1] {
	case 'p':
		port, perr := strconv.Atoi(suffix)
		if perr != nil {
			return false, nil
		}
		s.Port = port
		return true, nil
	case 'i':
		if suffix == "" {
			return false, nil
		}
		s.Identity = suffix
		return true, nil
	case 'F':
		if suffix == "" {
			return false, nil
		}
		s.ConfigFile = suffix
		return true, nil
	case 'l':
		if suffix == "" {
			return false, nil
		}
		s.User = suffix
		return true, nil
	case 'J':
		if suffix == "" {
			return false, nil
		}
		for _, j := range strings.Split(suffix, ",") {
			j = strings.TrimSpace(j)
			if j != "" {
				s.ProxyJump = append(s.ProxyJump, j)
			}
		}
		return true, nil
	case 'o':
		if suffix == "" {
			return false, nil
		}
		if err := mergeOption(s.Options, suffix); err != nil {
			return true, err
		}
		return true, nil
	default:
		return false, nil
	}
}

func mergeOption(opts map[string]string, val string) error {
	eq := strings.Index(val, "=")
	if eq < 0 {
		return fmt.Errorf("-o value must be key=value, got %q", val)
	}
	key := strings.ToLower(strings.TrimSpace(val[:eq]))
	opts[key] = strings.TrimSpace(val[eq+1:])
	return nil
}

func (s *Spec) applyLegacy() error {
	if s.LegacyHost != "" {
		if s.Destination != "" && s.Destination != s.LegacyHost {
			return fmt.Errorf("cannot use both positional destination and --host")
		}
		if s.Destination == "" {
			s.Destination = s.LegacyHost
		}
	}
	if s.LegacyUser != "" {
		if s.User != "" && s.User != s.LegacyUser {
			return fmt.Errorf("cannot use both -l/--user and --user")
		}
		s.User = s.LegacyUser
	}
	if s.LegacyPort >= 0 {
		if s.Port != 0 && s.Port != s.LegacyPort {
			return fmt.Errorf("cannot use both -p/--port and --port with different values")
		}
		s.Port = s.LegacyPort
	}
	if s.LegacyIdentity != "" {
		if s.Identity != "" && s.Identity != s.LegacyIdentity {
			return fmt.Errorf("cannot use both -i and --identity with different values")
		}
		s.Identity = s.LegacyIdentity
	}
	return nil
}

// ParseDestination splits [user@]host[:port] or ssh:// URI into user, host, port.
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
		end := strings.Index(dest, "]")
		if end < 0 {
			return "", "", 0, fmt.Errorf("invalid destination %q", dest)
		}
		hostPart = dest[1:end]
		rest := dest[end+1:]
		if strings.HasPrefix(rest, ":") {
			port, err = strconv.Atoi(rest[1:])
			if err != nil {
				return "", "", 0, fmt.Errorf("invalid port in %q", dest)
			}
		} else if rest != "" {
			return "", "", 0, fmt.Errorf("invalid destination %q", dest)
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
