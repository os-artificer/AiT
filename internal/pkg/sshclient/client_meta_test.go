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

package sshclient

import (
	"bytes"
	"context"
	"reflect"
	"testing"
)

func TestHandleMetaLineCLI(t *testing.T) {
	var got []string
	c := &Client{}
	err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, "!ait list", InteractiveOpts{
		OnMetaCLI: func(args []string) error {
			got = args
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, []string{"list"}) {
		t.Errorf("OnMetaCLI args = %v, want [list]", got)
	}
}

func TestHandleMetaLineHelp(t *testing.T) {
	var got []string
	c := &Client{}
	for _, line := range []string{"!ait help", "!ait"} {
		got = nil
		err := c.handleMetaLine(context.Background(), &bytes.Buffer{}, line, InteractiveOpts{
			OnMetaCLI: func(args []string) error {
				got = args
				return nil
			},
		})
		if err != nil {
			t.Fatalf("%q: %v", line, err)
		}
		if got != nil {
			t.Errorf("%q: want nil args, got %v", line, got)
		}
	}
}

func TestHandleMetaLineUnknownNoForward(t *testing.T) {
	c := &Client{}
	var stdin bytes.Buffer
	err := c.handleMetaLine(context.Background(), &stdin, "not-a-meta-line", InteractiveOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if stdin.Len() != 0 {
		t.Error("unknown line should not forward to stdin")
	}
}

func TestMetascannerHelpLine(t *testing.T) {
	fwd, _, meta := drive("!ait help\n", true)
	if len(fwd) != 0 {
		t.Errorf("forwarded %q", fwd)
	}
	if len(meta) != 1 || meta[0] != "!ait help" {
		t.Errorf("meta=%v want [!ait help]", meta)
	}
}
