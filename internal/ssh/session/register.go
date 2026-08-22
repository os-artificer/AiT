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

package session

import (
	"github.com/os-artificer/AiT/internal/ssh/registry"

	"github.com/spf13/cobra"
)

type topLevel struct{}

func init() {
	registry.Register(topLevel{})
}

func (topLevel) Name() string  { return Name }
func (topLevel) Short() string { return "list, load, suspend, switch, or delete saved sessions" }
func (topLevel) Order() int    { return 80 }
func (topLevel) NewCobra(d registry.Deps) *cobra.Command {
	return New(RunWithMgr(d.RunWithMgr))
}
