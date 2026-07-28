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

package logsrc

// Ring keeps the last capacity matching records (oldest first in Snapshot).
type Ring struct {
	buf      []Record
	capacity int
	full     bool
	pos      int
}

// NewRing returns a ring that retains at most capacity records.
// A capacity of zero or less makes Add a no-op and Snapshot return nil.
func NewRing(capacity int) *Ring {
	if capacity <= 0 {
		return &Ring{}
	}
	return &Ring{buf: make([]Record, capacity), capacity: capacity}
}

// Add appends rec, dropping the oldest record when the ring is full.
func (r *Ring) Add(rec Record) {
	if r.capacity == 0 {
		return
	}
	r.buf[r.pos] = rec
	r.pos++
	if r.pos >= r.capacity {
		r.pos = 0
		r.full = true
	}
}

// Snapshot returns the retained records in arrival order (oldest first).
func (r *Ring) Snapshot() []Record {
	if r.capacity == 0 {
		return nil
	}
	if !r.full {
		out := make([]Record, r.pos)
		copy(out, r.buf[:r.pos])
		return out
	}
	out := make([]Record, r.capacity)
	copy(out, r.buf[r.pos:])
	copy(out[r.capacity-r.pos:], r.buf[:r.pos])
	return out
}

// Len returns the number of retained records.
func (r *Ring) Len() int {
	if r.capacity == 0 {
		return 0
	}
	if r.full {
		return r.capacity
	}
	return r.pos
}
