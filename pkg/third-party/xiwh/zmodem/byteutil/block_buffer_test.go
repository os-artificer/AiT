package byteutil

import (
	"io"
	"testing"
	"time"
)

// TestBlockBuffer_CloseLargeSlowReader (V4) closes without close timeout while reader is slow.
func TestBlockBuffer_CloseLargeSlowReader(t *testing.T) {
	const size = 100 * 1024
	buf := NewBlockReadWriter(size)

	done := make(chan error, 1)
	go func() {
		_, err := io.Copy(io.Discard, buf)
		done <- err
	}()

	chunk := make([]byte, 32*1024)
	for written := 0; written < size; {
		n := copy(chunk, chunk)
		if written+n > size {
			n = size - written
		}
		if _, err := buf.Write(chunk[:n]); err != nil {
			t.Fatalf("Write: %v", err)
		}
		written += n
	}

	closeErr := make(chan error, 1)
	go func() {
		closeErr <- buf.Close()
	}()

	select {
	case err := <-closeErr:
		if err != nil {
			t.Fatalf("Close: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Close timed out")
	}

	select {
	case err := <-done:
		if err != nil && err != io.EOF {
			t.Fatalf("Copy: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reader did not finish")
	}
}
