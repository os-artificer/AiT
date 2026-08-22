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

package aitrelease

import (
	"time"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

const (
	defaultRepoOwner = "os-artificer"
	defaultRepoName  = "AiT"
	defaultRepoURL   = "https://github.com/os-artificer/AiT"

	defaultAPIBaseURL = "https://api.github.com"

	httpTimeout     = 2 * time.Minute
	executeTimeout  = 5 * time.Minute
	maxExtractBytes = 500 * 1024 * 1024

	pluginsDirName   = "plugins"
	pluginJSONName   = "plugin.json"
	installProbeName = ".ait-install-probe"

	aitBinaryName = toolname.AIT
)
