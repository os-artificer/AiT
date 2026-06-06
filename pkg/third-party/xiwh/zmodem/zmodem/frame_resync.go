package zmodem

const maxResyncSkip = 4096

// indexFrameStart returns the index of the first likely ZMODEM frame in data, or -1.
func indexFrameStart(data []byte) int {
	for i := 0; i+4 <= len(data); i++ {
		b0, b1, b2, b3 := data[i], data[i+1], data[i+2], data[i+3]
		if b0 == byte(ZPAD) && b1 == byte(ZPAD) && b2 == byte(ZDLE) && b3 == byte(ZHEX) {
			return i
		}
		if b0 == byte(ZPAD) && b1 == byte(ZDLE) {
			return i
		}
	}
	return -1
}

// resyncUnreadBuf drops leading non-frame bytes before the next frame marker.
// Caller must hold t.lock. Returns the number of bytes discarded.
func (t *ZModem) resyncUnreadBuf() int {
	idx := indexFrameStart(t.unreadBuf)
	if idx > 0 {
		if idx > maxResyncSkip {
			t.unreadBuf = t.unreadBuf[maxResyncSkip:]
			return maxResyncSkip
		}
		t.unreadBuf = t.unreadBuf[idx:]
		return idx
	}
	if idx == 0 {
		return 0
	}
	if len(t.unreadBuf) > maxResyncSkip {
		skipped := len(t.unreadBuf) - maxResyncSkip
		t.unreadBuf = t.unreadBuf[skipped:]
		return skipped
	}
	return 0
}

// PumpUnread resyncs buffered PTY data and starts a handler pass when data is pending.
func (t *ZModem) PumpUnread() {
	t.lock.Lock()
	status := t.status
	hasData := len(t.unreadBuf) > 0
	if hasData {
		t.resyncUnreadBuf()
	}
	t.lock.Unlock()

	if !hasData {
		return
	}
	if status == StatusReceive {
		go t.handleReceive()
	} else if status == StatusSend {
		go t.handleSend()
	}
}

// HasUnread reports whether the protocol buffer contains unread bytes.
func (t *ZModem) HasUnread() bool {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return len(t.unreadBuf) > 0
}

// EnterReceiveForTest puts the engine into receive state (tests only).
func (t *ZModem) EnterReceiveForTest() {
	t.lock.Lock()
	t.status = StatusReceive
	t.lock.Unlock()
}

// EnterSendForTest puts the engine into send state (tests only).
func (t *ZModem) EnterSendForTest() {
	t.lock.Lock()
	t.status = StatusSend
	t.lock.Unlock()
}
