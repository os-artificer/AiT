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

package procsig

import (
	"errors"
	"testing"
)

// AG-12: the wrapped error is returned unchanged so exit-code semantics survive.
func TestWithAbsorbedInterrupt_ReturnsFnError(t *testing.T) {
	want := errors.New("boom")
	if got := WithAbsorbedInterrupt(func() error { return want }); !errors.Is(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// AG-12: nested absorb windows must not leak the drain goroutine or handler.
func TestWithAbsorbedInterrupt_Nested(t *testing.T) {
	var inner bool
	err := WithAbsorbedInterrupt(func() error {
		return WithAbsorbedInterrupt(func() error {
			inner = true
			return nil
		})
	})
	if err != nil || !inner {
		t.Fatalf("err=%v inner=%v", err, inner)
	}
}

// A nil fn is a no-op rather than a panic.
func TestWithAbsorbedInterrupt_NilFn(t *testing.T) {
	if err := WithAbsorbedInterrupt(nil); err != nil {
		t.Fatalf("got %v, want nil", err)
	}
}
