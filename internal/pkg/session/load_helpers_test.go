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

package session

import (
	"errors"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/store"
)

func TestWrapLoadNotFound_DisplayIDHint(t *testing.T) {
	err := wrapLoadNotFound("621b71ac", store.ErrSessionNotFound)
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
	if !strings.Contains(err.Error(), "display only") {
		t.Fatalf("hint missing: %v", err)
	}
	if !strings.Contains(err.Error(), "Try:") {
		t.Fatalf("missing Try block: %v", err)
	}
}

func TestWrapLoadNotFound_AliasUnchanged(t *testing.T) {
	err := wrapLoadNotFound("ghost", store.ErrSessionNotFound)
	if !errors.Is(err, store.ErrSessionNotFound) {
		t.Fatalf("expected ErrSessionNotFound, got %v", err)
	}
	if strings.Contains(err.Error(), "Try:") {
		t.Fatal("non-display token should not get display hint")
	}
}

func TestIsListDisplayIDToken(t *testing.T) {
	if !isListDisplayIDToken("621b71ac") {
		t.Fatal("expected hex display token")
	}
	if isListDisplayIDToken("prod") {
		t.Fatal("alias should not match")
	}
	if isListDisplayIDToken("621b71") {
		t.Fatal("wrong length should not match")
	}
}
