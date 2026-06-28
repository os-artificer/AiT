package zmodem

import "bytes"

// SendFileEOF reports whether the upload data plane has finished (ZEOF sent).
func (t *ZModem) SendFileEOF() bool {
	return t.sendFileEOF
}

// ForceRelease ends the session and returns the engine to idle.
func (t *ZModem) ForceRelease() {
	t.release()
}

// TryCompleteSendTeardown parses a remote ZFIN or lrz "OO" ack after local ZFIN.
// Returns true when the engine was released to idle.
func (t *ZModem) TryCompleteSendTeardown() bool {
	t.lock.Lock()
	if !t.sendFileEOF || t.status != StatusSend {
		t.lock.Unlock()
		return false
	}

	t.resyncUnreadBuf()
	complete := false
	sendOO := false

	if idx := indexFrameStart(t.unreadBuf); idx >= 0 {
		f, n, err := unmarshalFrame(t.unreadBuf[idx:])
		if err == nil && f.frameType == ZFIN {
			t.unreadBuf = t.unreadBuf[idx+n:]
			complete = true
			sendOO = true
		}
	}

	if !complete && (bytes.HasSuffix(t.unreadBuf, []byte{'O', 'O'}) ||
		bytes.Contains(t.unreadBuf, []byte("OO\r\n")) ||
		bytes.Contains(t.unreadBuf, []byte("OO\n"))) {
		complete = true
	}

	t.lock.Unlock()
	if !complete {
		return false
	}

	if sendOO && t.consumer.Writer != nil {
		_, _ = t.consumer.Writer.Write([]byte{'O', 'O'})
	}
	t.release()
	return true
}
