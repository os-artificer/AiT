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

package helpdoc

import (
	"errors"
	"fmt"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

func TestIsUnknownCommand(t *testing.T) {
	if !IsUnknownCommand(errors.New(fmt.Sprintf(`unknown command "foo" for "%s"`, toolname.AISsh))) {
		t.Fatal("expected unknown command")
	}
	if IsUnknownCommand(errors.New("usage: session")) {
		t.Fatal("usage error is not unknown command")
	}
	if IsUnknownCommand(nil) {
		t.Fatal("nil is not unknown")
	}
}
