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

package hoststats

import (
	"strings"
	"testing"
)

func TestKillProc_InvalidPID(t *testing.T) {
	for _, pid := range []int32{0, -1, -99} {
		err := KillProc(pid, false)
		if err == nil {
			t.Fatalf("pid %d: want error", pid)
		}
		if !strings.Contains(err.Error(), "invalid pid") {
			t.Fatalf("pid %d: got %v", pid, err)
		}
	}
}

func TestKillProc_MissingProcess(t *testing.T) {
	// Extremely unlikely to exist; NewProcess or signal should fail.
	err := KillProc(1<<30, false)
	if err == nil {
		t.Fatal("want error for missing process")
	}
}
