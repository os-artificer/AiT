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

// Package all blank-imports every top-level command registrar (breaks import cycles with cli).
package all

import (
	_ "github.com/os-artificer/AiT/internal/ait/ai"
	_ "github.com/os-artificer/AiT/internal/ait/config"
	_ "github.com/os-artificer/AiT/internal/ait/get"
	_ "github.com/os-artificer/AiT/internal/ait/login"
	_ "github.com/os-artificer/AiT/internal/ait/put"
	_ "github.com/os-artificer/AiT/internal/ait/session"
	_ "github.com/os-artificer/AiT/internal/ait/shell"
	_ "github.com/os-artificer/AiT/internal/ait/version"
)
