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

package aitplugin

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Load reads and validates a Manifest from r (entire stream; trailing JSON is rejected).
func Load(r io.Reader) (Manifest, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Manifest{}, fmt.Errorf("read plugin.json: %w", err)
	}
	return LoadBytes(data)
}

// LoadBytes parses and validates JSON bytes as a Manifest.
func LoadBytes(data []byte) (Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("decode plugin.json: %w", err)
	}
	if err := Validate(&m); err != nil {
		return Manifest{}, err
	}
	return m, nil
}

// LoadFile reads path as plugin.json and validates it.
func LoadFile(path string) (Manifest, error) {
	f, err := os.Open(path)
	if err != nil {
		return Manifest{}, err
	}
	defer f.Close()
	return Load(f)
}
