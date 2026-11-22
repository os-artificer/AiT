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

package ocr_test

import (
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/vidsearch/frames"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/hit"
	"github.com/os-artificer/AiT/internal/pkg/vidsearch/ocr"
)

func TestSearch_SubstringHit(t *testing.T) {
	idx := ocr.Index{
		ByFrame: map[string]string{"/tmp/f.jpg": "plate 京A12345 here"},
		Frames:  []frames.Frame{{Path: "/tmp/f.jpg", TSec: 12}},
	}
	hits := ocr.Search(idx, "京A", hit.ModeText)
	if len(hits) != 1 {
		t.Fatalf("hits=%v", hits)
	}
	if hits[0].Kind != hit.KindText || hits[0].Score != 1 {
		t.Fatalf("%+v", hits[0])
	}
}

func TestSearch_NoMatch(t *testing.T) {
	idx := ocr.Index{
		ByFrame: map[string]string{"/tmp/f.jpg": "hello"},
		Frames:  []frames.Frame{{Path: "/tmp/f.jpg", TSec: 1}},
	}
	if hits := ocr.Search(idx, "missing", hit.ModeText); len(hits) != 0 {
		t.Fatalf("got %v", hits)
	}
}

func TestSearch_PlateRegex(t *testing.T) {
	idx := ocr.Index{
		ByFrame: map[string]string{"/tmp/f.jpg": "vehicle 京A12345 parked"},
		Frames:  []frames.Frame{{Path: "/tmp/f.jpg", TSec: 2}},
	}
	hits := ocr.Search(idx, "车牌", hit.ModeText)
	if len(hits) != 1 {
		t.Fatalf("hits=%v", hits)
	}
	if hits[0].Kind != hit.KindRegex {
		t.Fatalf("kind=%s", hits[0].Kind)
	}
	if hits[0].Snippet == "" {
		t.Fatal("empty snippet")
	}
}

func TestSearch_PhoneRegex(t *testing.T) {
	idx := ocr.Index{
		ByFrame: map[string]string{"/tmp/f.jpg": "call 138-1234-5678 now"},
		Frames:  []frames.Frame{{Path: "/tmp/f.jpg", TSec: 3}},
	}
	hits := ocr.Search(idx, "电话", hit.ModeText)
	if len(hits) != 1 {
		t.Fatalf("hits=%v", hits)
	}
	if hits[0].Kind != hit.KindRegex {
		t.Fatalf("kind=%s", hits[0].Kind)
	}
}
