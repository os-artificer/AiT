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
	"github.com/os-artificer/AiT/internal/pkg/cliroute"
	"github.com/os-artificer/AiT/internal/ssh/meta"
	"github.com/os-artificer/AiT/internal/ssh/registry"
	_ "github.com/os-artificer/AiT/internal/ssh/registry/all"
	"github.com/os-artificer/AiT/internal/ssh/shell"
)

func init() {
	names := make([]string, 0, len(registry.Entries()))
	for _, e := range registry.Entries() {
		names = append(names, e.Name)
	}
	cliroute.SetReservedSubcommandNames(names)

	meta.SetExecutorWriters(ExecuteCobraOnly)
	shell.SetCLIExecutor(ExecuteCobraOnly)
}
