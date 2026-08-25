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

package aissh

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/clihints"
	"github.com/os-artificer/AiT/internal/pkg/config"
	pkghelpdoc "github.com/os-artificer/AiT/internal/pkg/helpdoc"
	"github.com/os-artificer/AiT/internal/pkg/session"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
	"github.com/os-artificer/AiT/internal/ssh/meta"
	"github.com/os-artificer/AiT/internal/ssh/shell"
)

// OA20: aissh no longer registers an ai subtree.
func TestAissh_AISubtreeRemoved(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	err := ExecuteWriters([]string{"ai"}, io.Discard, io.Discard)
	if !pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("want unknown command, got %v", err)
	}
}

// OA22: !ait ai in SSH meta delegates to unknown + Try toward ai binary.
func TestExecute_AIMetaUnknownHint(t *testing.T) {
	meta.SetExecutorWriters(func(args []string, _, _ io.Writer) error {
		return fmt.Errorf(`unknown command "%s" for "%s"`, args[0], toolname.AISsh)
	})
	err := meta.Execute([]string{"ai"})
	if err == nil {
		t.Fatal("expected error for ai meta subcommand")
	}
	if !pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("want unknown command, got %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("expected Try block: %v", err)
	}
	if !strings.Contains(err.Error(), clihints.TryAIBinary) {
		t.Fatalf("expected Try ai binary, got %v", err)
	}
}

// OA23: aissh shell no longer treats ai as a built-in verb.
func TestShell_AINotBuiltin(t *testing.T) {
	t.Setenv("AIT_HOME", t.TempDir())
	shell.SetCLIExecutor(func(args []string, _, _ io.Writer) error {
		return fmt.Errorf(`unknown command "%s" for "%s"`, args[0], toolname.AISsh)
	})

	_, err := shell.ExecuteLine(context.Background(), &session.Manager{Config: &config.Config{}}, "ai", nil)
	if err == nil {
		t.Fatal("expected error for shell ai command")
	}
	if !pkghelpdoc.IsUnknownCommand(err) {
		t.Fatalf("want unknown command via CLI dispatch, got %v", err)
	}
}

// OA24: internal/ssh must not import internal/pkg/aimode (AI lives in opsai plugin).
func TestAissh_NoAimodeImport(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "ssh")
	const aimodePkg = "github.com/os-artificer/AiT/internal/pkg/aimode"
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			p := strings.Trim(imp.Path.Value, `"`)
			if p == aimodePkg || strings.HasPrefix(p, aimodePkg+"/") {
				t.Errorf("%s imports %s", path, p)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// OA25: aissh ExecuteWriters no longer strips --ai-experimental or flips the CLI gate.
func TestAissh_NoExperimentalStrip(t *testing.T) {
	config.SetAIExperimentalCLI(false)
	t.Cleanup(func() { config.SetAIExperimentalCLI(false) })
	t.Setenv("AIT_HOME", t.TempDir())
	t.Setenv(config.EnvAIExperimental, "")

	err := ExecuteWriters([]string{config.FlagAIExperimental, "session", "--list"}, io.Discard, io.Discard)
	if err == nil {
		t.Fatal("expected error for unknown --ai-experimental flag")
	}
	if config.AIExperimentalCLI() {
		t.Fatal("ExecuteWriters must not SetAIExperimentalCLI")
	}
}
