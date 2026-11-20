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

package config

import "testing"

func TestSetGetKey_AIVisionModel(t *testing.T) {
	cfg := &Config{}
	if err := cfg.SetKey(keyAIVisionModel, "gpt-4o"); err != nil {
		t.Fatal(err)
	}
	if cfg.AI.Vision.Model != "gpt-4o" {
		t.Fatalf("model=%q", cfg.AI.Vision.Model)
	}
	got, err := cfg.GetKey(keyAIVisionModel)
	if err != nil || got != "gpt-4o" {
		t.Fatalf("GetKey=%q err=%v", got, err)
	}
}
