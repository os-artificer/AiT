package zmodem

import (
	"io"
	"testing"
)

// TestTryCompleteSendTeardown_OOAck releases on lrz OO echo after local ZFIN.
func TestTryCompleteSendTeardown_OOAck(t *testing.T) {
	var ptyOut capturePTY
	z := New(ZModemConsumer{
		Writer:     &ptyOut,
		EchoWriter: io.Discard,
		OnUpload: func() *ZModemFile {
			return nil
		},
	})
	z.lock.Lock()
	z.status = StatusSend
	z.sendFileEOF = true
	z.unreadBuf = append(z.unreadBuf, []byte("Transfer complete\r\nOO\r\n")...)
	z.lock.Unlock()

	if !z.TryCompleteSendTeardown() {
		t.Fatal("expected OO ack to complete send teardown")
	}
	if z.GetStatus() != StatusIdle {
		t.Fatalf("status = %v, want idle", z.GetStatus())
	}
}
