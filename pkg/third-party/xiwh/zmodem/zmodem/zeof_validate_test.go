package zmodem

import (
	"encoding/binary"
	"io"
	"testing"

	"github.com/os-artificer/AiT/pkg/third-party/xiwh/zmodem/byteutil"
)

// TestZModemReceive_ZEOFOffsetMismatch (V3) rejects ZEOF when header offset != bytes received.
func TestZModemReceive_ZEOFOffsetMismatch(t *testing.T) {
	buf := byteutil.NewBlockReadWriter(5)
	if _, err := buf.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := buf.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(io.Discard, buf); err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if buf.BytesReceived() != 5 {
		t.Fatalf("BytesReceived=%d want 5", buf.BytesReceived())
	}

	hdr := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdr, 99)
	if zeofOffsetMatches(hdr, buf) {
		t.Fatal("expected mismatch for wrong offset")
	}

	binary.LittleEndian.PutUint32(hdr, 5)
	if !zeofOffsetMatches(hdr, buf) {
		t.Fatal("expected match for correct offset")
	}
}
