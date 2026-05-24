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

package cli

import (
	"fmt"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
)

func unknownProviderErr(id string) error {
	return errhint.WithTryMsg(fmt.Sprintf("unknown provider %q", id),
		"Use deepseek, minimax, or ollama.",
		"ait ai provider enable deepseek",
		"ait ai provider enable minimax",
		"ait ai provider enable ollama",
		"ait ai provider --help",
	)
}
