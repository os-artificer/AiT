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

package client

import (
	"net"
	"sync/atomic"
	"testing"
)

// trackingConn records whether Close was called (PJ1).
type trackingConn struct {
	net.Conn
	closed atomic.Bool
}

func (t *trackingConn) Close() error {
	t.closed.Store(true)
	if t.Conn != nil {
		return t.Conn.Close()
	}
	return nil
}

// TestDialProxyJumpChain_KeepsJumpAlive (PJ1) verifies attaching via clients and
// that Close tears down jump hops only after the final client Close path runs —
// the regression was closing the jump client before the final handshake.
func TestDialProxyJumpChain_KeepsJumpAlive(t *testing.T) {
	a, b := net.Pipe()
	t.Cleanup(func() {
		_ = a.Close()
		_ = b.Close()
	})

	jumpAgent := &trackingConn{Conn: a}
	jump := &Client{agentConn: jumpAgent}
	finalAgent := &trackingConn{Conn: b}
	final := &Client{agentConn: finalAgent, via: []*Client{jump}}

	if jumpAgent.closed.Load() {
		t.Fatal("jump must stay open while final client is live")
	}

	if err := final.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if !finalAgent.closed.Load() {
		t.Fatal("final agentConn should be closed")
	}
	if !jumpAgent.closed.Load() {
		t.Fatal("jump via client should close after final Close")
	}
}

// TestClient_CloseClosesNestedViaChain (PJ1) closes via hops innermost-first.
func TestClient_CloseClosesNestedViaChain(t *testing.T) {
	outerA, outerB := net.Pipe()
	innerA, innerB := net.Pipe()
	t.Cleanup(func() {
		_ = outerA.Close()
		_ = outerB.Close()
		_ = innerA.Close()
		_ = innerB.Close()
	})

	var order []string
	innerTrack := &orderConn{Conn: innerA, name: "inner", order: &order}
	outerTrack := &orderConn{Conn: outerA, name: "outer", order: &order}
	finalTrack := &orderConn{Conn: outerB, name: "final", order: &order}

	inner := &Client{agentConn: innerTrack}
	outer := &Client{agentConn: outerTrack, via: []*Client{inner}}
	final := &Client{agentConn: finalTrack, via: []*Client{outer}}

	if err := final.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	want := []string{"final", "outer", "inner"}
	if len(order) != len(want) {
		t.Fatalf("close order %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("close order %v, want %v", order, want)
		}
	}
}

type orderConn struct {
	net.Conn
	name  string
	order *[]string
}

func (o *orderConn) Close() error {
	*o.order = append(*o.order, o.name)
	if o.Conn != nil {
		return o.Conn.Close()
	}
	return nil
}
