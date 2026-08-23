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

package pluginbin_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/pluginbin"
)

func TestExecPath_UsesRunner(t *testing.T) {
	t.Cleanup(func() { pluginbin.SetRunner(nil) })

	called := false
	pluginbin.SetRunner(func(_ context.Context, path string, argv []string) error {
		called = true
		if path != "/bin/tool" || argv[0] != "tool" {
			t.Fatalf("path=%q argv=%v", path, argv)
		}
		return nil
	})

	err := pluginbin.ExecPath(context.Background(), "/bin/tool", []string{"tool", "--help"}, nil, io.Discard, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("runner not called")
	}
}

func TestExecPath_RunnerError(t *testing.T) {
	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	want := errors.New("boom")
	pluginbin.SetRunner(func(context.Context, string, []string) error { return want })
	if err := pluginbin.ExecPath(context.Background(), "p", []string{"p"}, nil, nil, nil); !errors.Is(err, want) {
		t.Fatalf("err = %v", err)
	}
}

func TestSetRunner_ReturnsPrevious(t *testing.T) {
	t.Cleanup(func() { pluginbin.SetRunner(nil) })
	first := pluginbin.Runner(func(context.Context, string, []string) error { return nil })
	prev := pluginbin.SetRunner(first)
	if prev != nil {
		t.Fatal("expected nil previous")
	}
	second := pluginbin.Runner(func(context.Context, string, []string) error { return errors.New("x") })
	got := pluginbin.SetRunner(second)
	if got == nil {
		t.Fatal("expected previous runner")
	}
}
