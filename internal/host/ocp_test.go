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

package aihost_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AH15: command packages must not import sibling command packages.
func TestCommandPackages_NoSiblingImports(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	siblings := []string{"cpu", "mem", "disk", "net", "proto", "proc", "snapshot", "sources", "top", "version"}
	sibSet := map[string]bool{}
	for _, s := range siblings {
		sibSet[s] = true
	}
	for _, name := range siblings {
		dir := filepath.Join(wd, name)
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		fset := token.NewFileSet()
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") || strings.HasSuffix(e.Name(), "_test.go") {
				continue
			}
			path := filepath.Join(dir, e.Name())
			f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatal(err)
			}
			for _, imp := range f.Imports {
				p := strings.Trim(imp.Path.Value, `"`)
				for sib := range sibSet {
					if sib == name {
						continue
					}
					if strings.HasSuffix(p, "/host/"+sib) {
						t.Fatalf("%s imports sibling %s via %s", name, sib, p)
					}
				}
			}
		}
	}
}

func TestRegistryAll_BlankImportsCommands(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(wd, "registry", "all", "imports.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"/cpu", "/mem", "/disk", "/net", "/proto", "/proc", "/top", "/snapshot", "/sources", "/version"} {
		if !strings.Contains(text, want) {
			t.Fatalf("imports.go missing blank-import for %s", want)
		}
	}
}
