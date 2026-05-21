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

package args

import (
	"fmt"
	"strconv"
	"strings"
)

// applyFlag consumes one flag at argv[i] and returns the index of the next token.
// cmdLabel names the subcommand in error text (e.g. login, get, put).
func applyFlag(s *Spec, argv []string, i int, cmdLabel string) (int, error) {
	arg := argv[i]
	if msg, ok := unsupportedFlags[arg]; ok {
		return i, fmt.Errorf("%s is not supported by ait %s; use the system ssh client", msg, cmdLabel)
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

// applyGluedFlag handles OpenSSH glued forms like -p22 or -i~/.ssh/id_ed25519.
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

func applyLegacy(s *Spec) error {
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
