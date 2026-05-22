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

package aimode

import (
	"testing"
)

func TestIsMetaREPLCommand(t *testing.T) {
	meta := []string{"help", "model list", "model", "clear", "policy once", "run echo", "ait list", "-h", "--help"}
	for _, line := range meta {
		if !isMetaREPLCommand(line) {
			t.Errorf("%q should be meta", line)
		}
	}
	nonMeta := []string{"how do I grep logs", "what is docker", "chat hello world", ""}
	for _, line := range nonMeta {
		if isMetaREPLCommand(line) {
			t.Errorf("%q should not be meta", line)
		}
	}
}
