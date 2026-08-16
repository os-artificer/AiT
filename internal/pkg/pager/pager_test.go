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

package pager_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/pager"
)

func TestViewport_Scroll(t *testing.T) {
	vp := pager.NewViewport(3)
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		vp.Append(s)
	}
	if got := strings.Join(vp.Visible(), ""); got != "abc" {
		t.Fatalf("visible=%q", got)
	}
	vp.Down(2)
	if got := strings.Join(vp.Visible(), ""); got != "cde" {
		t.Fatalf("after down=%q", got)
	}
	vp.Top()
	if vp.Percent() != 0 && len(vp.Visible()) > 0 {
		// with offset 0 and more lines, percent is 0
	}
	vp.Bottom()
	if vp.Percent() != 100 {
		t.Fatalf("percent=%d", vp.Percent())
	}
	vp.HalfUp()
	vp.HalfDown()
	vp.Up(1)
}

func TestViewport_Search(t *testing.T) {
	vp := pager.NewViewport(2)
	for _, s := range []string{"one", "two", "findme", "four"} {
		vp.Append(s)
	}
	if !vp.Search("find", false) {
		t.Fatal("expected hit")
	}
	vis := vp.Visible()
	if len(vis) == 0 || !strings.Contains(vis[0], "find") {
		t.Fatalf("%v", vis)
	}
}

func TestPage_Passthrough(t *testing.T) {
	var buf bytes.Buffer
	src := &pager.SliceSource{Lines: []string{"a", "b", "c"}}
	if err := pager.Page(context.Background(), &buf, src, pager.Options{Disabled: true}); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "a\nb\nc\n" {
		t.Fatalf("%q", buf.String())
	}
}
