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

// Package all blank-imports every top-level aithost command registrar.
package all

import (
	_ "github.com/os-artificer/AiT/internal/aithost/cpu"
	_ "github.com/os-artificer/AiT/internal/aithost/disk"
	_ "github.com/os-artificer/AiT/internal/aithost/mem"
	_ "github.com/os-artificer/AiT/internal/aithost/net"
	_ "github.com/os-artificer/AiT/internal/aithost/proc"
	_ "github.com/os-artificer/AiT/internal/aithost/snapshot"
	_ "github.com/os-artificer/AiT/internal/aithost/sources"
	_ "github.com/os-artificer/AiT/internal/aithost/top"
	_ "github.com/os-artificer/AiT/internal/aithost/version"
)
