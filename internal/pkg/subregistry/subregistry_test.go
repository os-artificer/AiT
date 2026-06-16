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

package subregistry

import (
	"strings"
	"testing"
)

type testCmd struct {
	name  string
	short string
	order int
}

func (c testCmd) Name() string  { return c.name }
func (c testCmd) Short() string { return c.short }
func (c testCmd) Order() int    { return c.order }

func TestRegistry_RegisterSortEntries(t *testing.T) {
	var _ MetaCommand = testCmd{}

	reg := New[testCmd]("test")
	reg.Register(testCmd{name: "b", short: "B", order: 20})
	reg.Register(testCmd{name: "a", short: "A", order: 10})
	reg.Register(testCmd{name: "c", short: "C", order: 20})

	entries := reg.Entries()
	if len(entries) != 3 {
		t.Fatalf("entries=%d", len(entries))
	}
	if entries[0].Name != "a" || entries[1].Name != "b" || entries[2].Name != "c" {
		t.Fatalf("order=%v", entries)
	}

	lines := FormatSubcommandLines(entries)
	if !strings.Contains(lines, "  a   A\n") || !strings.Contains(lines, "  b   B\n") {
		t.Fatalf("lines=%q", lines)
	}
	if JoinChildNames(entries, "|") != "a|b|c" {
		t.Fatalf("join=%q", JoinChildNames(entries, "|"))
	}
}

func TestRegistry_DuplicatePanics(t *testing.T) {
	reg := New[testCmd]("dup")
	reg.Register(testCmd{name: "x", short: "X"})
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	reg.Register(testCmd{name: "x", short: "Y"})
}
