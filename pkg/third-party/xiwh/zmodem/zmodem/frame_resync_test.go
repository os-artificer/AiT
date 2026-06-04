package zmodem

import (
	"io"
	"testing"
	"time"
)

// TestZModem_ResyncSkipsPrefix (V11) drops printable PTY noise before ZFIN frames.
func TestZModem_ResyncSkipsPrefix(t *testing.T) {
	zfin, err := newHexFrame(ZFIN, DEFAULT_HEADER_DATA).marshal()
	if err != nil {
		t.Fatal(err)
	}

	prefix := []byte("Bytes/sec 12345\r\nTransfer complete\r\n")
	var ptyOut capturePTY
	z := New(ZModemConsumer{
		Writer:     &ptyOut,
		EchoWriter: io.Discard,
		OnUpload: func() *ZModemFile {
			return nil
		},
		OnUploadSkip:    func(*ZModemFile) {},
		OnCheckDownload: func(*ZModemFile) {},
		OnDownload: func(*ZModemFile, io.ReadCloser) error {
			return nil
		},
	})

	z.lock.Lock()
	z.status = StatusReceive
	z.unreadBuf = append(append([]byte(nil), prefix...), zfin...)
	z.lock.Unlock()

	z.PumpUnread()

	deadline := time.Now().Add(2 * time.Second)
	for z.GetStatus() != StatusIdle {
		if time.Now().After(deadline) {
			t.Fatalf("status still %v after resync pump", z.GetStatus())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

type capturePTY []byte

func (c *capturePTY) Write(p []byte) (int, error) {
	*c = append(*c, p...)
	return len(p), nil
}
