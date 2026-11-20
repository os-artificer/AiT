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

// Package hit defines the shared video search Hit type (leaf; no vidsearch imports).
package hit

// Hit is one search match in a video.
type Hit struct {
	Video     string  `json:"video"`
	VideoName string  `json:"video_name"`
	CameraID  string  `json:"camera_id"`
	TSec      float64 `json:"t_sec"`
	TEndSec   float64 `json:"t_end_sec,omitempty"`
	Frame     string  `json:"frame,omitempty"`
	Sidecar   string  `json:"sidecar,omitempty"`
	Kind      string  `json:"kind"`
	Query     string  `json:"query"`
	Snippet   string  `json:"snippet"`
	Score     float64 `json:"score"`
	Mode      string  `json:"mode"`
}

// Kind and mode string constants.
const (
	KindText   = "text"
	KindObject = "object"
	KindRegex  = "regex"

	ModeText   = "text"
	ModeObject = "object"
	ModeAuto   = "auto"
)
