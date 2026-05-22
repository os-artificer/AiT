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
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/config"
)

// ChooseModelsDir resolves and persists models_dir before install/pull/serve.
// flagDir overrides interactive selection when non-empty.
func ChooseModelsDir(in io.Reader, out io.Writer, cfg *config.Config, assumeYes bool, flagDir string) (string, error) {
	dir, err := ResolveModelsDir(cfg, flagDir)
	if err != nil {
		return "", err
	}
	if flagDir != "" || assumeYes {
		return persistModelsDir(cfg, out, dir)
	}
	if existing := ModelsDirFromConfig(cfg); existing != "" {
		return persistModelsDir(cfg, out, existing)
	}

	opts, err := SuggestedModelsDirs(cfg)
	if err != nil {
		return "", err
	}
	if !isInteractive(in) {
		return persistModelsDir(cfg, out, dir)
	}

	_, _ = fmt.Fprintln(out, "Choose where Ollama stores downloaded models (OLLAMA_MODELS):")
	for i, o := range opts {
		_, _ = fmt.Fprintf(out, "  %d) %s\n     %s\n", i+1, o.Label, o.Path)
	}
	customN := len(opts) + 1
	_, _ = fmt.Fprintf(out, "  %d) Enter a custom absolute path\n", customN)
	_, _ = fmt.Fprint(out, "Selection [1]: ")

	choice, err := readChoiceLine(in)
	if err != nil {
		return "", err
	}
	if choice == "" {
		choice = "1"
	}

	switch choice {
	case "1", "2":
		idx := 0
		if choice == "2" {
			idx = 1
		}
		if idx >= len(opts) {
			return "", fmt.Errorf("invalid selection %q", choice)
		}
		dir = opts[idx].Path
	case fmt.Sprintf("%d", customN):
		_, _ = fmt.Fprint(out, "Absolute path: ")
		custom, err := readChoiceLine(in)
		if err != nil {
			return "", err
		}
		dir, err = expandModelsPath(strings.TrimSpace(custom))
		if err != nil {
			return "", err
		}
	default:
		if n, err := parseChoiceIndex(choice); err == nil && n >= 1 && n <= len(opts) {
			dir = opts[n-1].Path
		} else if n == customN {
			_, _ = fmt.Fprint(out, "Absolute path: ")
			custom, err := readChoiceLine(in)
			if err != nil {
				return "", err
			}
			dir, err = expandModelsPath(strings.TrimSpace(custom))
			if err != nil {
				return "", err
			}
		} else {
			return "", fmt.Errorf("invalid selection %q", choice)
		}
	}

	return persistModelsDir(cfg, out, dir)
}

func persistModelsDir(cfg *config.Config, out io.Writer, dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("create models directory: %w", err)
	}
	if err := SetProviderModelsDir(cfg, dir); err != nil {
		return "", err
	}
	if out != nil {
		_, _ = fmt.Fprintf(out, "Models directory (OLLAMA_MODELS): %s\n", dir)
	}
	return dir, nil
}

func isInteractive(in io.Reader) bool {
	if in == nil {
		return false
	}
	f, ok := in.(*os.File)
	if !ok {
		return false
	}
	return f == os.Stdin && isTerminalIn(f)
}

func readChoiceLine(in io.Reader) (string, error) {
	br, ok := in.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(in)
	}
	line, err := br.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func parseChoiceIndex(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n, err
}
