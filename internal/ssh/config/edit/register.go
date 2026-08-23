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

package edit

import (
	"github.com/os-artificer/AiT/internal/ssh/config/subregistry"

	"github.com/spf13/cobra"
)

type command struct{}

func init() {
	subregistry.Register(command{})
}

func (command) Name() string  { return name }
func (command) Short() string { return "Edit config.yaml in $VISUAL or $EDITOR" }
func (command) Order() int    { return 30 }
func (command) NewCobra() *cobra.Command {
	return New()
}
