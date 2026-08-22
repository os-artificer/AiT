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

package put

import "github.com/os-artificer/AiT/internal/pkg/ssh/openssh"

// TransferSpec holds parsed argv for aissh put.
type TransferSpec = openssh.TransferSpec

// ParseTransfer parses argv for aissh put.
func ParseTransfer(argv []string) (TransferSpec, error) {
	return openssh.ParseTransfer(argv, openssh.OpPut)
}
