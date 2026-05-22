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

package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEnsureModel_SkipWhenPresent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"models":[{"name":"llama3.2"}]}`))
	}))
	defer srv.Close()

	if err := EnsureModel(context.Background(), srv.Client(), srv.URL+"/v1", "llama3.2", false, ""); err != nil {
		t.Fatal(err)
	}
}

func TestEnsureModel_SkipPullFlag(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"models":[]}`))
	}))
	defer srv.Close()

	if err := EnsureModel(context.Background(), srv.Client(), srv.URL+"/v1", "llama3.2", true, ""); err != nil {
		t.Fatal(err)
	}
}
