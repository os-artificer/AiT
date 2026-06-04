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
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"
)

type stallReader struct {
	mu     sync.Mutex
	rel    sync.Cond
	data   []byte
	closed bool
}

func newStallReader() *stallReader {
	s := &stallReader{}
	s.rel = *sync.NewCond(&s.mu)
	return s
}

func (s *stallReader) unblock() {
	s.mu.Lock()
	s.closed = true
	s.rel.Broadcast()
	s.mu.Unlock()
}

func (s *stallReader) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for len(s.data) == 0 && !s.closed {
		s.rel.Wait()
	}
	if len(s.data) == 0 {
		return 0, io.EOF
	}
	n := copy(p, s.data)
	s.data = s.data[n:]
	return n, nil
}

// TestCopyWithContext_CancelDuringBlockedRead (V9) exits soon after ctx cancel on blocked Read.
func TestCopyWithContext_CancelDuringBlockedRead(t *testing.T) {
	stall := newStallReader()
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		err := copyWithContext(ctx, io.Discard, newCtxReader(ctx, stall), 0, nil)
		errCh <- err
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		stall.unblock()
		t.Fatal("copyWithContext did not cancel promptly")
	}
}
