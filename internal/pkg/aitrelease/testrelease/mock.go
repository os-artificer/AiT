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

package testrelease

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ServeOptions configures a long-running mock GitHub releases/latest server.
type ServeOptions struct {
	Tag         string
	ArchivePath string
	GOOS        string
	GOARCH      string
}

// ListenAndServe starts an HTTP server that mimics GitHub releases/latest and asset download.
// baseURL is suitable for AIT_RELEASE_TEST_BASE_URL. shutdown stops the server gracefully.
func ListenAndServe(addr string, opts ServeOptions) (baseURL string, shutdown func(context.Context) error, err error) {
	if strings.TrimSpace(opts.Tag) == "" {
		return "", nil, fmt.Errorf("tag is required")
	}
	if addr == "" {
		addr = ":8765"
	}

	assetName, archiveBytes, err := loadReleaseArchive(opts)
	if err != nil {
		return "", nil, err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", nil, err
	}

	mux := http.NewServeMux()
	srv := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	base := "http://" + ln.Addr().String()
	mux.HandleFunc("/", releaseHandler(base, opts.Tag, assetName, archiveBytes))

	go func() {
		_ = srv.Serve(ln)
	}()

	return base, srv.Shutdown, nil
}

func loadReleaseArchive(opts ServeOptions) (assetName string, data []byte, err error) {
	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = runtime.GOARCH
	}
	assetName = AssetFileName(opts.Tag, goos, goarch)

	if opts.ArchivePath != "" {
		data, err = os.ReadFile(opts.ArchivePath)
		if err != nil {
			return "", nil, err
		}
		return assetName, data, nil
	}

	if strings.HasSuffix(assetName, ".zip") {
		return "", nil, fmt.Errorf("stub bundle on windows requires -archive (zip)")
	}

	dir, err := os.MkdirTemp("", "ait-testrelease-*")
	if err != nil {
		return "", nil, err
	}
	defer os.RemoveAll(dir)

	if err := WriteBundleTree(dir, "aissh-bin", "ailog-bin"); err != nil {
		return "", nil, err
	}

	archivePath := filepath.Join(os.TempDir(), "ait-testrelease-"+assetName)
	if err := PackTarGzDir(dir, archivePath); err != nil {
		return "", nil, err
	}
	defer os.Remove(archivePath)
	data, err = os.ReadFile(archivePath)
	if err != nil {
		return "", nil, err
	}
	return assetName, data, nil
}

func releaseHandler(baseURL, tag, assetName string, archiveBytes []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			payload, _ := json.Marshal(map[string]any{
				"tag_name": tag,
				"assets": []map[string]string{{
					"name":                 assetName,
					"browser_download_url": baseURL + "/assets/" + assetName,
				}},
			})
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(payload)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			_, _ = w.Write(archiveBytes)
			return
		}
		http.NotFound(w, r)
	}
}
