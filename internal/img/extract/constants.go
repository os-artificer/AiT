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

// Package extract implements the default aiimg route: metadata extraction for image files.
// It is not registered as a Cobra subcommand (package name ≠ CLI verb).
package extract

import (
	"time"

	"github.com/os-artificer/AiT/internal/pkg/imgmeta"
)

// ErrPartialFailure is an alias of imgmeta.ErrPartialFailure for callers that already import extract.
var ErrPartialFailure = imgmeta.ErrPartialFailure

const (
	geocodeMinInterval = time.Second
	projectURL         = "https://github.com/os-artificer/AiT"
)
