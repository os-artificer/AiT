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

package ollama

import (
	"fmt"
	"io"
	"runtime"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// ManualSetupInstructions prints install and config steps when auto-install was declined.
func ManualSetupInstructions(w io.Writer, cfg *config.Config) {
	if w == nil {
		return
	}
	model := config.DefaultOllamaModel
	base := config.DefaultOllamaBase
	if cfg != nil {
		if pc, ok := cfg.Provider(config.ProviderOllama); ok {
			if pc.Model != "" {
				model = pc.Model
			}
			if pc.BaseURL != "" {
				base = pc.BaseURL
			}
		}
	}

	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Manual Ollama setup:")
	switch runtime.GOOS {
	case "darwin":
		_, _ = fmt.Fprintln(w, "  brew install ollama")
	case "linux":
		_, _ = fmt.Fprintf(w, "  curl -fsSL %s | sh\n", installScriptURL)
		_, _ = fmt.Fprintln(w, "  (may require sudo)")
	case "windows":
		_, _ = fmt.Fprintf(w, "  winget install Ollama.Ollama\n  or download: %s\n", windowsDownloadURL)
	default:
		_, _ = fmt.Fprintln(w, "  https://ollama.com/download")
	}
	_, _ = fmt.Fprintln(w, "  ollama serve")
	_, _ = fmt.Fprintf(w, "  ollama pull %s\n", model)
	_, _ = fmt.Fprintln(w, "  Or run: "+toolname.AICLI("ollama setup"))
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Config snippet (config.yaml):")
	_, _ = fmt.Fprintln(w, "  ai:")
	_, _ = fmt.Fprintln(w, "    tiered:")
	_, _ = fmt.Fprintln(w, "      enabled: true")
	_, _ = fmt.Fprintln(w, "    providers:")
	_, _ = fmt.Fprintln(w, "      ollama:")
	modelsDir := ""
	if cfg != nil {
		modelsDir = ModelsDirFromConfig(cfg)
	}
	_, _ = fmt.Fprintf(w, "        enabled: true\n        base_url: %s\n        model: %s\n", base, model)
	if modelsDir != "" {
		_, _ = fmt.Fprintf(w, "        models_dir: %s\n", modelsDir)
		_, _ = fmt.Fprintln(w, "")
		_, _ = fmt.Fprintln(w, "  If Ollama runs as a system service, set the same path:")
		_, _ = fmt.Fprintf(w, "    export OLLAMA_MODELS=%s\n", modelsDir)
	}
}
