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

package testhttp

import (
	"net"
	"net/http"
	"testing"
	"time"
)

// Start listens on an ephemeral loopback port and registers t.Cleanup.
func Start(t *testing.T, cfg Config) *Server {
	t.Helper()
	srv, err := Listen(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv
}

// Listen starts the mock HTTP server. The caller must Close the returned Server.
func Listen(cfg Config) (*Server, error) {
	cfg = applyHTTPDefaults(cfg)
	addr := cfg.Addr
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", serveScenario(cfg))
	hs := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	srv := &Server{cfg: cfg, ln: ln, http: hs, done: make(chan struct{})}
	go func() {
		defer close(srv.done)
		_ = hs.Serve(ln)
	}()
	return srv, nil
}

func applyHTTPDefaults(cfg Config) Config {
	if cfg.Scenario == "" {
		cfg.Scenario = ScenarioOK
	}
	if cfg.ChatReply == "" {
		cfg.ChatReply = defaultChatReply
	}
	if cfg.DisplayName == "" {
		cfg.DisplayName = defaultDisplayName
	}
	if cfg.OllamaModel == "" {
		cfg.OllamaModel = defaultOllamaModel
	}
	return cfg
}
