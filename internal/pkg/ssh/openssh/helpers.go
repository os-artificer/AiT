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

func nextFlagValue(argv []string, i int, flag string) (int, string, error) {
	i++
	if i >= len(argv) {
		return i, "", fmt.Errorf("%s requires a value", flag)
	}
	return i, argv[i], nil
}

func setPortFromString(s *Spec, flag, val string) error {
	port, err := strconv.Atoi(val)
	if err != nil {
		return fmt.Errorf("invalid %s: %w", flag, err)
	}
	s.Port = port
	return nil
}

func appendProxyJumpFromString(s *Spec, val string) {
	for j := range strings.SplitSeq(val, ",") {
		j = strings.TrimSpace(j)
		if j != "" {
			s.ProxyJump = append(s.ProxyJump, j)
		}
	}
}

func applyGluedPort(s *Spec, suffix string) (bool, error) {
	port, err := strconv.Atoi(suffix)
	if err != nil {
		return false, nil
	}
	s.Port = port
	return true, nil
}

func applyGluedString(set func(string), suffix string) (bool, error) {
	if suffix == "" {
		return false, nil
	}
	set(suffix)
	return true, nil
}
