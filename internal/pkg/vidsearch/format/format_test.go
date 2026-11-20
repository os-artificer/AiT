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

package format_test

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/format"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
)

func TestWriteHuman_GroupsAndSummary(t *testing.T) {
	var buf bytes.Buffer
	err := format.WriteHuman(&buf, []format.VideoResult{
		{
			Video:    "/abs/clip1.mp4",
			CameraID: "cam03",
			Hits: []hit.Hit{{
				TSec: 12, Kind: hit.KindText, Snippet: "京A12345", Score: 1, Frame: "./aitvid-hits/cam03/x.jpg",
			}},
		},
		{Video: "/abs/clip2.mp4", CameraID: "", Hits: nil},
	}, format.Summary{Videos: 2, Hits: 1, Failed: 0, OutDir: "./aitvid-hits"})
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	for _, want := range []string{
		"# /abs/clip1.mp4  camera=cam03",
		`  12.0s  text   "京A12345"  score=1.00`,
		"# /abs/clip2.mp4  camera=-",
		"(no hits)",
		"2 videos, 1 hits, 0 failed; saved ./aitvid-hits",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
}

func TestWriteNDJSON_Fields(t *testing.T) {
	var buf bytes.Buffer
	h := hit.Hit{
		Video: "/abs/clip1.mp4", VideoName: "clip1.mp4", CameraID: "cam03",
		TSec: 12, Frame: "/abs/f.jpg", Sidecar: "/abs/f.json",
		Kind: hit.KindText, Query: "京A", Snippet: "京A12345", Score: 1, Mode: hit.ModeText,
	}
	if err := format.WriteNDJSON(&buf, []hit.Hit{h}); err != nil {
		t.Fatal(err)
	}
	var got hit.Hit
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Video != h.Video || got.CameraID != "cam03" || got.Kind != hit.KindText {
		t.Fatalf("got %+v", got)
	}
}

func TestWriteNDJSON_Empty(t *testing.T) {
	var buf bytes.Buffer
	if err := format.WriteNDJSON(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.Len() != 0 {
		t.Fatalf("want empty stdout, got %q", buf.String())
	}
}
