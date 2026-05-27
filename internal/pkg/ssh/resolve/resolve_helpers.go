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

package resolve

import (
	"bufio"
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/platform"
)

type sshGConfig struct {
	single map[string]string
	multi  map[string][]string
}

func parseSSHGOutput(out []byte) sshGConfig {
	cfg := sshGConfig{
		single: make(map[string]string),
		multi:  make(map[string][]string),
	}
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := strings.ToLower(fields[0])
		val := strings.TrimSpace(strings.Join(fields[1:], " "))
		cfg.single[key] = val
		cfg.multi[key] = append(cfg.multi[key], val)
	}
	return cfg
}

func expandIdentityPaths(paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}
		out = append(out, platform.ExpandSSHPath(p))
	}
	return out
}

func firstExistingIdentity(paths []string) string {
	for _, p := range paths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func parsePort(s string) int {
	if s == "" {
		return 0
	}
	p, _ := strconv.Atoi(s)
	return p
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
