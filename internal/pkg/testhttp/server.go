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
	"sync"
	"time"
)

// Server is a loopback HTTP mock for OpenAI, Ollama, and Nominatim.
type Server struct {
	cfg       Config
	ln        net.Listener
	http      *http.Server
	done      chan struct{}
	closeOnce sync.Once
}

// Close stops the HTTP server.
func (s *Server) Close() error {
	var err error
	s.closeOnce.Do(func() {
		err = s.http.Close()
		<-s.done
	})
	return err
}

// BaseURL is the origin, e.g. http://127.0.0.1:PORT.
func (s *Server) BaseURL() string {
	return "http://" + s.ln.Addr().String()
}

// ChatURL is the OpenAI-compatible /v1 base (append /chat/completions).
func (s *Server) ChatURL() string {
	return s.BaseURL() + "/v1"
}

// ReverseURL is the Nominatim reverse endpoint.
func (s *Server) ReverseURL() string {
	return s.BaseURL() + pathReverse
}

// HTTPClient returns a client with a short timeout suitable for mock calls.
func (s *Server) HTTPClient() *http.Client {
	return &http.Client{Timeout: 8 * time.Second}
}

// Info returns fixture metadata for pytest / cmd/testenv-mock.
func (s *Server) Info() Info {
	return Info{
		BaseURL:     s.BaseURL(),
		ChatURL:     s.ChatURL(),
		TagsURL:     s.BaseURL() + pathAPITags,
		ReverseURL:  s.ReverseURL(),
		Scenario:    s.cfg.Scenario,
		ChatReply:   s.cfg.ChatReply,
		DisplayName: s.cfg.DisplayName,
	}
}
