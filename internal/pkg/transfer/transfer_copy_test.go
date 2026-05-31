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
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
)

func TestCopyWithContext_ReportsProgress(t *testing.T) {
	src := strings.NewReader("hello world")
	var dst bytes.Buffer
	var calls []int64
	progress := CopyProgress(func(written, total int64) {
		calls = append(calls, written)
	})

	if err := copyWithContext(context.Background(), &dst, src, int64(src.Len()), progress); err != nil {
		t.Fatal(err)
	}
	if len(calls) == 0 {
		t.Fatal("expected progress callbacks")
	}
	if calls[len(calls)-1] != int64(len("hello world")) {
		t.Fatalf("last written = %d, want %d", calls[len(calls)-1], len("hello world"))
	}
}

func TestCopyWithContext_NilProgress(t *testing.T) {
	src := strings.NewReader("x")
	var dst bytes.Buffer
	if err := copyWithContext(context.Background(), &dst, src, 1, nil); err != nil {
		t.Fatal(err)
	}
}

func TestCopyWithContext_ContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	src := io.MultiReader(strings.NewReader("abc"))
	var dst bytes.Buffer
	err := copyWithContext(ctx, &dst, src, 3, nil)
	if err != context.Canceled {
		t.Fatalf("expected canceled, got %v", err)
	}
}
