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

package ollama

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveInstallAllowed_YesFlag(t *testing.T) {
	allowed, err := ResolveInstallAllowed(true, Options{AssumeYes: true}, nil)
	if err != nil || !allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}

func TestResolveInstallAllowed_NoInstall(t *testing.T) {
	allowed, err := ResolveInstallAllowed(true, Options{NoInstall: true}, nil)
	if err != nil || allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}

func TestResolveInstallAllowed_ConfirmN(t *testing.T) {
	confirm := func(string) (bool, error) { return false, nil }
	allowed, err := ResolveInstallAllowed(true, Options{}, confirm)
	if err != nil || allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}

func TestResolveInstallAllowed_ConfirmY(t *testing.T) {
	confirm := func(prompt string) (bool, error) {
		if !strings.Contains(prompt, "[y/N]") {
			t.Fatalf("prompt=%q", prompt)
		}
		return true, nil
	}
	allowed, err := ResolveInstallAllowed(true, Options{Out: &bytes.Buffer{}}, confirm)
	if err != nil || !allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}

func TestResolveInstallAllowed_BinaryPresent(t *testing.T) {
	allowed, err := ResolveInstallAllowed(false, Options{}, nil)
	if err != nil || allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}
