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

package get

import "github.com/os-artificer/AiT/internal/pkg/ssh/openssh"

// OpenSSH-compatible and AiT transfer flags accepted by ai-ssh get (see internal/pkg/ssh/openssh).
const (
	FlagSession  = openssh.FlagSession
	FlagPassword = openssh.FlagPassword
	FlagHost     = openssh.FlagHost
	FlagUser     = openssh.FlagUser
	FlagPort     = openssh.FlagPort
	FlagIdentity = openssh.FlagIdentity
)
