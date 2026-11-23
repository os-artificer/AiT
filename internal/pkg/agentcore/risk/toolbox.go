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

package risk

import (
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// ClassifyToolbox returns the risk Level for a toolbox argv line
// (e.g. "aihost snapshot", "ailog search error").
func ClassifyToolbox(argsLine string) Level {
	fields := strings.Fields(strings.TrimSpace(argsLine))
	if len(fields) == 0 {
		return Low
	}
	bin := strings.ToLower(fields[0])
	if canon, ok := aitplugin.LegacyPluginNames[bin]; ok {
		bin = canon
	}
	rest := ""
	if len(fields) > 1 {
		rest = strings.ToLower(strings.Join(fields[1:], " "))
	}

	switch bin {
	case toolname.Host, toolname.AITHost:
		return classifyAIHost(rest, fields[1:])
	case toolname.Log, toolname.AITLog:
		return classifyAILog(rest, fields[1:])
	case toolname.Img, toolname.AITImg:
		return Read
	case toolname.Vid, toolname.AITVid:
		return classifyAIVid(rest, fields[1:])
	case toolname.SSH, toolname.AITSSH:
		return classifyAisshToolbox(rest)
	default:
		return Low
	}
}

func classifyAIVid(rest string, args []string) Level {
	_ = rest
	if len(args) == 0 {
		return Read
	}
	sub := strings.ToLower(args[0])
	switch sub {
	case "deps":
		for _, a := range args[1:] {
			if a == "--install" {
				return Low // Write: package install; Low forces confirm (not Read auto-run)
			}
		}
		return Read
	case "frames", "version":
		return Read
	case "search":
		for _, a := range args[1:] {
			if a == "--no-save-hits" {
				return Read
			}
		}
		// Default --save-hits writes JPEGs under --out-dir.
		return Low
	default:
		return Low
	}
}

func classifyAisshToolbox(rest string) Level {
	lower := strings.ToLower(strings.TrimSpace(rest))
	if strings.Contains(lower, "session") && strings.Contains(lower, "--list") {
		return Read
	}
	return ClassifyAit(rest)
}

func classifyAIHost(rest string, args []string) Level {
	_ = rest
	if len(args) == 0 {
		return Interactive // default top
	}
	sub := strings.ToLower(args[0])
	switch sub {
	case "top":
		return Interactive
	case "snapshot", "cpu", "mem", "disk", "net", "proc", "proto", "sources", "version":
		return Read
	default:
		return Low
	}
}

func classifyAILog(rest string, args []string) Level {
	joined := rest
	for _, a := range args {
		if a == "-f" || a == "--follow" || a == "follow" {
			return Interactive
		}
	}
	if strings.Contains(joined, "--follow") || strings.Contains(" "+joined+" ", " -f ") {
		return Interactive
	}
	if len(args) == 0 {
		return Read
	}
	sub := strings.ToLower(args[0])
	switch sub {
	case "search", "sources", "version", "view":
		return Read
	default:
		return Read
	}
}
