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

package geocode_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/geocode"
)

func TestNominatim_Reverse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("format") != "jsonv2" {
			t.Errorf("format=%q", r.URL.Query().Get("format"))
		}
		if !strings.Contains(r.Header.Get("Accept-Language"), "zh-CN") {
			t.Errorf("Accept-Language=%q", r.Header.Get("Accept-Language"))
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		_, _ = w.Write([]byte(`{"display_name":"上海市黄浦区外滩, 上海, 中国"}`))
	}))
	defer srv.Close()

	c := &geocode.Client{BaseURL: srv.URL, HTTP: srv.Client(), UserAgent: "test"}
	addr, err := c.Reverse(context.Background(), 31.24, 121.49)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(addr, "外滩") {
		t.Fatalf("addr=%q", addr)
	}
}

func TestNominatim_ReverseError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()

	c := &geocode.Client{BaseURL: srv.URL, HTTP: srv.Client()}
	_, err := c.Reverse(context.Background(), 1, 2)
	if err == nil {
		t.Fatal("expected error")
	}
}
