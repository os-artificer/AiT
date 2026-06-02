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
		var val string
		var err error
		i, val, err = nextFlagValue(argv, i, "-p")
		if err != nil {
			return i, err
		}
		if err := setPortFromString(s, "-p", val); err != nil {
			return i, err
		}
		return i, nil
	case "-l":
		return applyStringFlag(s, argv, i, "-l", func(v string) { s.User = v })
	case "-i":
		return applyStringFlag(s, argv, i, "-i", func(v string) { s.Identity = v })
	case "-F":
		return applyStringFlag(s, argv, i, "-F", func(v string) { s.ConfigFile = v })
	case "-J":
		var val string
		var err error
		i, val, err = nextFlagValue(argv, i, "-J")
		if err != nil {
			return i, err
		}
		appendProxyJumpFromString(s, val)
		return i, nil
	case "-o":
		var val string
		var err error
		i, val, err = nextFlagValue(argv, i, "-o")
		if err != nil {
			return i, err
		}
		if err := mergeOption(s.Options, val); err != nil {
			return i, err
		}
		return i, nil
	case "--alias":
		return applyStringFlag(s, argv, i, "--alias", func(v string) { s.Alias = v })
	case "--save":
		s.Save = true
		return i, nil
	case "--tmux":
		s.Tmux = true
		return i, nil
	case "--password":
		return applyStringFlag(s, argv, i, "--password", func(v string) { s.Password = v })
	case "--host":
		return applyStringFlag(s, argv, i, "--host", func(v string) { s.LegacyHost = v })
	case "--user":
		return applyStringFlag(s, argv, i, "--user", func(v string) { s.LegacyUser = v })
	case "--port":
		var val string
		var err error
		i, val, err = nextFlagValue(argv, i, "--port")
		if err != nil {
			return i, err
		}
		port, err := strconv.Atoi(val)
		if err != nil {
			return i, fmt.Errorf("invalid --port: %w", err)
		}
		s.LegacyPort = port
		return i, nil
	case "--identity":
		return applyStringFlag(s, argv, i, "--identity", func(v string) { s.LegacyIdentity = v })
	default:
		if strings.HasPrefix(arg, "-o") && len(arg) > 2 {
			if err := mergeOption(s.Options, arg[2:]); err != nil {
				return i, err
			}
			return i, nil
		}
		return i, unknownFlagErr(arg, cmdLabel)
	}
}

func applyStringFlag(s *Spec, argv []string, i int, flag string, set func(string)) (int, error) {
	var val string
	var err error
	i, val, err = nextFlagValue(argv, i, flag)
	if err != nil {
		return i, err
	}
	set(val)
	return i, nil
}

// applyGluedFlag handles OpenSSH glued forms like -p22 or -i~/.ssh/id_ed25519.
func applyGluedFlag(s *Spec, arg string) (glued bool, err error) {
	if len(arg) < 3 || arg[0] != '-' {
		return false, nil
	}
	suffix := arg[2:]
	switch arg[1] {
	case 'p':
		return applyGluedPort(s, suffix)
	case 'i':
		return applyGluedString(func(v string) { s.Identity = v }, suffix)
	case 'F':
		return applyGluedString(func(v string) { s.ConfigFile = v }, suffix)
	case 'l':
		return applyGluedString(func(v string) { s.User = v }, suffix)
	case 'J':
		if suffix == "" {
			return false, nil
		}
		appendProxyJumpFromString(s, suffix)
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
