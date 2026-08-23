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

package init

import (
	"github.com/os-artificer/AiT/internal/ssh/config/subregistry"

	"github.com/spf13/cobra"
)

type command struct{}

func init() {
	subregistry.Register(command{})
}

func (command) Name() string  { return name }
func (command) Short() string { return "Create or update config.yaml with defaults for missing keys" }
func (command) Order() int    { return 10 }
func (command) NewCobra() *cobra.Command {
	return New()
}
