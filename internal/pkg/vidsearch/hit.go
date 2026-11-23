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

// Package vidsearch searches video files for text and object content.
package vidsearch

import "github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"

// Hit is one search match in a video.
type Hit = hit.Hit

// Kind and mode string constants.
const (
	KindText   = hit.KindText
	KindObject = hit.KindObject
	KindRegex  = hit.KindRegex

	ModeText   = hit.ModeText
	ModeObject = hit.ModeObject
	ModeAuto   = hit.ModeAuto
)
