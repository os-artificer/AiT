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

package aitrelease

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// ReleaseAsset is one file attached to a GitHub release.
type ReleaseAsset struct {
	Name               string
	BrowserDownloadURL string
}

// Release is a GitHub release payload subset.
type Release struct {
	TagName string
	Assets  []ReleaseAsset
}

type githubRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// FetchLatestRelease returns the latest non-draft GitHub release for the configured repo.
func FetchLatestRelease(ctx context.Context, client *http.Client, baseURL, owner, repo, userAgent string) (Release, error) {
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}
	if baseURL == "" {
		baseURL = defaultAPIBaseURL
	}
	if owner == "" {
		owner = defaultRepoOwner
	}
	if repo == "" {
		repo = defaultRepoName
	}

	url := strings.TrimRight(baseURL, "/") + "/repos/" + owner + "/" + repo + "/releases/latest"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return Release{}, err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if token := strings.TrimSpace(os.Getenv("AIT_GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return Release{}, ErrNoRelease
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return Release{}, fmt.Errorf("github releases/latest: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var raw githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return Release{}, err
	}
	if raw.TagName == "" {
		return Release{}, ErrNoRelease
	}

	out := Release{TagName: raw.TagName}
	for _, a := range raw.Assets {
		out.Assets = append(out.Assets, ReleaseAsset{
			Name:               a.Name,
			BrowserDownloadURL: a.BrowserDownloadURL,
		})
	}
	return out, nil
}

// SelectAsset picks the platform archive name for tag from release assets.
func SelectAsset(release Release, goos, goarch string) (ReleaseAsset, error) {
	if goos == "" {
		goos = runtime.GOOS
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	want := assetFileName(release.TagName, goos, goarch)
	for _, a := range release.Assets {
		if a.Name == want {
			return a, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("%w: want %q", ErrNoAsset, want)
}

func assetFileName(tag, goos, goarch string) string {
	if goos == "windows" {
		return fmt.Sprintf("ait-%s-windows-amd64.zip", tag)
	}
	return fmt.Sprintf("ait-%s-%s-%s.tar.gz", tag, goos, goarch)
}

// DownloadAsset streams a release asset to a local file path.
func DownloadAsset(ctx context.Context, client *http.Client, asset ReleaseAsset, destPath, userAgent string) error {
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}
	if asset.BrowserDownloadURL == "" {
		return fmt.Errorf("%w: empty download url", ErrBundleInvalid)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}
	if userAgent != "" {
		req.Header.Set("User-Agent", userAgent)
	}
	if token := strings.TrimSpace(os.Getenv("AIT_GITHUB_TOKEN")); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: %s", asset.Name, resp.Status)
	}

	f, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
