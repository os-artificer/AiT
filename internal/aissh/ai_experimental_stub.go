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
	"os"

	"github.com/os-artificer/AiT/internal/pkg/config"
	"github.com/os-artificer/AiT/internal/pkg/errhint"
	sshclient "github.com/os-artificer/AiT/internal/pkg/ssh/client"
)

// disabledRemoteAIRunner is injected when AI experimental features are off.
// It always prints the disabled hint (does not re-evaluate CLI/env/config) and
// returns nil so the SSH session stays up.
func disabledRemoteAIRunner(_ sshclient.RemoteAIParams) error {
	errhint.Print(os.Stderr, errhint.SurfaceTTY, config.AIExperimentalDisabledErr())
	return nil
}
