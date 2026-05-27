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

package transfer

import (
	"testing"

	sshparse "github.com/os-artificer/AiT/internal/pkg/ssh/parse"
)

func TestParseTransferViaArgsPackage(t *testing.T) {
	spec, err := sshparse.ParseTransfer([]string{"-p22", "user@h:/r", "./l"}, sshparse.OpGet)
	if err != nil {
		t.Fatal(err)
	}
	if spec.Conn.Port != 22 || spec.RemotePath != "/r" || spec.LocalPath != "./l" {
		t.Fatalf("spec=%+v", spec)
	}
}
