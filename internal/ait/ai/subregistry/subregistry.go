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

// Package subregistry collects ait ai subcommands registered from init hooks.
package subregistry

import (
	"fmt"
	"sort"

	"github.com/os-artificer/AiT/internal/ait/ai/deps"

	"github.com/spf13/cobra"
)

// Command is the extension point for an `ait ai` subcommand.
type Command interface {
	Name() string
	Short() string
	Order() int
	NewCobra(d deps.Deps) *cobra.Command
}

// Entry is name and short description for parent help.
type Entry struct {
	Name  string
	Short string
	Order int
}

var registered []Command

// Register adds an ait ai subcommand. Panics on duplicate Name.
func Register(c Command) {
	for _, existing := range registered {
		if existing.Name() == c.Name() {
			panic(fmt.Sprintf("ai/subregistry: duplicate subcommand %q", c.Name()))
		}
	}
	registered = append(registered, c)
}

// Attach adds all registered subcommands to parent in Order ascending.
func Attach(parent *cobra.Command, d deps.Deps) {
	cmds := append([]Command(nil), registered...)
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Order() != cmds[j].Order() {
			return cmds[i].Order() < cmds[j].Order()
		}
		return cmds[i].Name() < cmds[j].Name()
	})
	for _, c := range cmds {
		parent.AddCommand(c.NewCobra(d))
	}
}

// Entries returns registered subcommands for ait ai --help (sorted by Order).
func Entries() []Entry {
	cmds := append([]Command(nil), registered...)
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
