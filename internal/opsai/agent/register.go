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

package agent

import (
	"github.com/os-artificer/AiT/internal/opsai/deps"
	"github.com/os-artificer/AiT/internal/opsai/subregistry"

	"github.com/spf13/cobra"
)

type command struct{}

func init() {
	subregistry.Register(command{})
}

func (command) Name() string { return name }
func (command) Short() string {
	return "Send one message to the AI in agent mode and execute tool calls (non-interactive)"
}
func (command) Order() int { return 20 }
func (command) NewCobra(d deps.Deps) *cobra.Command {
	return New(d)
}
