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

// Package registry collects top-level aissh subcommands registered from init hooks.
package registry

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

var registered []TopLevel

// Register adds a top-level command. Panics if Name duplicates an existing registration.
func Register(cmd TopLevel) {
	for _, existing := range registered {
		if existing.Name() == cmd.Name() {
			panic(fmt.Sprintf("registry: duplicate top-level command %q", cmd.Name()))
		}
	}
	registered = append(registered, cmd)
}

// AttachRoot adds all registered commands to root in Order ascending.
func AttachRoot(root *cobra.Command, d Deps) {
	cmds := append([]TopLevel(nil), registered...)
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Order() != cmds[j].Order() {
			return cmds[i].Order() < cmds[j].Order()
		}
		return cmds[i].Name() < cmds[j].Name()
	})
	for _, c := range cmds {
		root.AddCommand(c.NewCobra(d))
	}
}

// Entries returns registered commands for root --help (sorted by Order).
func Entries() []Entry {
	cmds := append([]TopLevel(nil), registered...)
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Order() != cmds[j].Order() {
			return cmds[i].Order() < cmds[j].Order()
		}
		return cmds[i].Name() < cmds[j].Name()
	})
	out := make([]Entry, len(cmds))
	for i, c := range cmds {
		out[i] = Entry{Name: c.Name(), Short: c.Short(), Order: c.Order()}
	}
	return out
}
