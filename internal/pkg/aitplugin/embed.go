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
	"embed"
	"fmt"
	"io"
)

//go:embed manifests
var manifestsFS embed.FS

// WriteManifest writes the embedded plugin.json for binary to w.
// binary must be a shipped tool name such as aissh.
// Returns an error when binary is unknown or the embed is missing.
func WriteManifest(w io.Writer, binary string) error {
	data, err := BuiltinJSON(binary)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// BuiltinJSON returns the embedded plugin.json bytes for binary.
func BuiltinJSON(binary string) ([]byte, error) {
	path := "manifests/" + binary + "/" + pluginJSONName
	data, err := manifestsFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("embedded manifest for %s: %w", binary, err)
	}
	return data, nil
}
