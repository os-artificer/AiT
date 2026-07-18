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

package connect

import (
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/ssh/resolve"
)

func TestOptsFromResolvedSave(t *testing.T) {
	opts := OptsFromResolved(resolve.Resolved{Host: "h", Save: true, Alias: "prod"})
	if !opts.Save {
		t.Fatal("Save not set")
	}
	if opts.Alias != "prod" {
		t.Fatalf("alias = %q", opts.Alias)
	}
}
