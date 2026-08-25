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

// Package testrelease provides httptest fixtures for aitrelease integration tests.
package testrelease

import (
	"archive/tar"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// NewServer returns an httptest.Server that serves a fake GitHub latest release and asset archive.
func NewServer(t *testing.T, tag string) *httptest.Server {
	t.Helper()

	assetName, archiveBytes, err := loadReleaseArchive(ServeOptions{Tag: tag})
	if err != nil {
		if strings.Contains(err.Error(), "stub bundle on windows") {
			t.Skip(err.Error())
		}
		t.Fatal(err)
	}

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		releaseHandler(srv.URL, tag, assetName, archiveBytes)(w, r)
	}))
	return srv
}

// AssetFileName returns the release archive name for the given tag and platform.
func AssetFileName(tag, goos, goarch string) string {
	if goos == "windows" {
		return "ait-" + tag + "-windows-amd64.zip"
	}
	return "ait-" + tag + "-" + goos + "-" + goarch + ".tar.gz"
}

// WriteBundleTree writes a minimal flat-root release bundle into dir.
func WriteBundleTree(dir, aisshBody, ailogBody string) error {
	bodies := map[string]string{
		"ai":     "ai-bin",
		"aissh":  aisshBody,
		"ailog":  ailogBody,
		"aiimg":  "aiimg-bin",
		"aihost": "aihost-bin",
	}
	for tool, body := range bodies {
		if err := writeBundleTool(dir, tool, body); err != nil {
			return err
		}
	}
	return os.WriteFile(filepath.Join(dir, toolname.AIT), []byte("ait-bin"), 0o755)
}

func writeBundleTool(dir, tool, body string) error {
	if err := os.WriteFile(filepath.Join(dir, tool), []byte(body), 0o755); err != nil {
		return err
	}
	plug := filepath.Join(dir, "plugins", tool)
	if err := os.MkdirAll(plug, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(plug, tool), []byte(body), 0o755); err != nil {
		return err
	}
	data, err := aitplugin.BuiltinJSON(tool)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(plug, "plugin.json"), data, 0o644)
}

// PackTarGzDir creates a tar.gz archive from a flat directory tree.
func PackTarGzDir(srcDir, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := tw.WriteHeader(&tar.Header{Name: rel, Mode: 0o755, Size: int64(len(data))}); err != nil {
			return err
		}
		_, err = tw.Write(data)
		return err
	})
	if err != nil {
		f.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		f.Close()
		return err
	}
	if err := gw.Close(); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
