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

// Package aitplugin defines the AiT toolbox plugin manifest schema and discovery.
package aitplugin

// Manifest is the v1 plugin descriptor used by ait for help and exec routing.
type Manifest struct {
	APIVersion int      `json:"apiVersion"`
	Name       string   `json:"name"`
	Binary     string   `json:"binary"`
	Short      string   `json:"short"`
	Order      int      `json:"order"`
	Usage      string   `json:"usage"`
	Flags      []Flag   `json:"flags,omitempty"`
	Examples   []string `json:"examples,omitempty"`
}

// Flag documents a plugin CLI flag for generated help only (ait does not parse it).
type Flag struct {
	Name  string `json:"name"`
	Short string `json:"short,omitempty"`
	Type  string `json:"type"`
	Help  string `json:"help"`
}

// Plugin is a validated manifest paired with a resolved executable path.
type Plugin struct {
	Manifest Manifest
	Path     string
}
