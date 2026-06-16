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

// Package subregistry provides a generic Cobra subcommand registry for nested CLI parents.
package subregistry

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"
)

// MetaCommand is the metadata shared by nested subcommands (name, short, order).
type MetaCommand interface {
	Name() string
	Short() string
	Order() int
}

// Entry is name and short description for parent --help.
type Entry struct {
	Name  string
	Short string
	Order int
}

// Registry collects subcommands registered from init hooks.
type Registry[C MetaCommand] struct {
	label string
	cmds  []C
}

// New returns an empty registry. label is used in duplicate-name panic messages.
func New[C MetaCommand](label string) *Registry[C] {
	return &Registry[C]{label: label}
}

// Register adds a subcommand. Panics on duplicate Name.
func (r *Registry[C]) Register(c C) {
	for _, existing := range r.cmds {
		if existing.Name() == c.Name() {
			panic(fmt.Sprintf("%s: duplicate subcommand %q", r.label, c.Name()))
		}
	}
	r.cmds = append(r.cmds, c)
}

// Attach adds all registered subcommands to parent in Order ascending.
func (r *Registry[C]) Attach(parent *cobra.Command, newCmd func(C) *cobra.Command) {
	for _, c := range r.sorted() {
		parent.AddCommand(newCmd(c))
	}
}

// Entries returns registered subcommands for parent --help (sorted by Order).
func (r *Registry[C]) Entries() []Entry {
	cmds := r.sorted()
	out := make([]Entry, len(cmds))
	for i, c := range cmds {
		out[i] = Entry{Name: c.Name(), Short: c.Short(), Order: c.Order()}
	}
	return out
}

func (r *Registry[C]) sorted() []C {
	cmds := append([]C(nil), r.cmds...)
	sort.Slice(cmds, func(i, j int) bool {
		if cmds[i].Order() != cmds[j].Order() {
			return cmds[i].Order() < cmds[j].Order()
		}
		return cmds[i].Name() < cmds[j].Name()
	})
	return cmds
}
