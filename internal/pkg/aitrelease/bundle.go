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
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/os-artificer/AiT/internal/pkg/aitplugin"
	"github.com/os-artificer/AiT/internal/pkg/shellext"
	"github.com/os-artificer/AiT/internal/pkg/toolname"
)

// Action selects install, update, or upgrade behavior.
type Action int

const (
	ActionInstall Action = iota
	ActionUpdate
	ActionUpgrade
)

// Options configures GitHub fetch and install layout.
type Options struct {
	HTTPClient  *http.Client
	BaseURL     string
	RepoOwner   string
	RepoName    string
	InstallRoot string
	UserAgent   string
}

// Request is one install/update/upgrade invocation.
type Request struct {
	Action Action
	Names  []string
}

// Plan describes what will be applied from a release bundle.
type Plan struct {
	Tools      []string
	IncludeAIT bool
	Skipped    []string
}

// BuildPlan resolves target tools for req without network I/O.
func BuildPlan(ctx context.Context, req Request) (Plan, error) {
	switch req.Action {
	case ActionInstall:
		return buildInstallPlan(req.Names)
	case ActionUpdate:
		return buildUpdatePlan(ctx, req.Names)
	case ActionUpgrade:
		if len(req.Names) > 0 {
			return Plan{}, ErrUpgradeExtraArgs
		}
		tools := aitplugin.InstalledOfficialPlugins(ctx)
		return Plan{Tools: tools, IncludeAIT: true}, nil
	default:
		return Plan{}, fmt.Errorf("unknown action")
	}
}

func buildInstallPlan(names []string) (Plan, error) {
	if len(names) == 0 {
		return Plan{Tools: aitplugin.OfficialPluginNames()}, nil
	}
	tools := make([]string, 0, len(names))
	for _, name := range names {
		if !aitplugin.IsOfficialPlugin(name) {
			return Plan{}, fmt.Errorf("%w: %s", ErrUnknownPlugin, name)
		}
		tools = append(tools, name)
	}
	sort.Strings(tools)
	return Plan{Tools: tools}, nil
}

func buildUpdatePlan(ctx context.Context, names []string) (Plan, error) {
	installed := aitplugin.InstalledOfficialPlugins(ctx)
	if len(names) == 0 {
		if len(installed) == 0 {
			return Plan{}, ErrNothingToUpdate
		}
		return Plan{Tools: installed}, nil
	}

	installedSet := make(map[string]struct{}, len(installed))
	for _, name := range installed {
		installedSet[name] = struct{}{}
	}

	var plan Plan
	for _, name := range names {
		if !aitplugin.IsOfficialPlugin(name) {
			return Plan{}, fmt.Errorf("%w: %s", ErrUnknownPlugin, name)
		}
		if _, ok := installedSet[name]; ok {
			plan.Tools = append(plan.Tools, name)
			continue
		}
		plan.Skipped = append(plan.Skipped, name)
	}
	sort.Strings(plan.Tools)
	sort.Strings(plan.Skipped)
	if len(plan.Tools) == 0 {
		return Plan{}, ErrNothingToUpdate
	}
	return plan, nil
}

// Execute fetches the latest release, extracts the platform bundle, and applies plan.
func Execute(ctx context.Context, opts Options, req Request, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}

	plan, err := BuildPlan(ctx, req)
	if err != nil {
		return err
	}

	installRoot, err := resolveInstallRoot(opts.InstallRoot)
	if err != nil {
		return err
	}
	if err := ensureInstallRootWritable(installRoot); err != nil {
		return fmt.Errorf("%w: %v", ErrInstallRoot, err)
	}

	ctx, cancel := context.WithTimeout(ctx, executeTimeout)
	defer cancel()

	client := opts.HTTPClient
	if client == nil {
		client = &http.Client{Timeout: httpTimeout}
	}

	release, err := FetchLatestRelease(ctx, client, opts.BaseURL, opts.RepoOwner, opts.RepoName, opts.UserAgent)
	if err != nil {
		return err
	}

	asset, err := SelectAsset(release, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "fetching %s...\n", asset.Name)

	tmpDir, err := os.MkdirTemp("", "ait-release-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, asset.Name)
	if err := DownloadAsset(ctx, client, asset, archivePath, opts.UserAgent); err != nil {
		return err
	}

	extractDir := filepath.Join(tmpDir, "extract")
	if err := extractArchive(archivePath, extractDir); err != nil {
		return fmt.Errorf("%w: %v", ErrBundleInvalid, err)
	}

	if err := applyBundle(installRoot, extractDir, plan); err != nil {
		return err
	}

	for _, name := range plan.Skipped {
		_, _ = fmt.Fprintf(out, "skipped %s: not installed (try: install %s)\n", name, name)
	}

	aitplugin.InvalidateDiscover()

	printApplyResults(out, req.Action, plan, release.TagName)

	return nil
}

func printApplyResults(out io.Writer, action Action, plan Plan, tag string) {
	verb := "updated"
	if action == ActionInstall {
		verb = "installed"
	}
	if plan.IncludeAIT {
		_, _ = fmt.Fprintf(out, "upgrade complete: restart ait to use %s\n", tag)
		if len(plan.Tools) == 0 {
			_, _ = fmt.Fprintf(out, "updated ait -> %s\n", tag)
		}
	}
	for _, tool := range plan.Tools {
		_, _ = fmt.Fprintf(out, "%s %s -> %s\n", verb, tool, tag)
	}
}

func resolveInstallRoot(override string) (string, error) {
	if override != "" {
		return filepath.Abs(override)
	}
	root, err := aitplugin.PrimaryInstallDir()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInstallRoot, err)
	}
	return root, nil
}

func ensureInstallRootWritable(root string) error {
	if err := os.MkdirAll(filepath.Join(root, pluginsDirName), 0o755); err != nil {
		return err
	}
	probe := filepath.Join(root, installProbeName)
	if err := os.WriteFile(probe, []byte("ok"), 0o600); err != nil {
		return err
	}
	return os.Remove(probe)
}

func extractArchive(archivePath, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return err
	}
	if strings.HasSuffix(strings.ToLower(archivePath), ".zip") {
		return extractZip(archivePath, destDir)
	}
	return extractTarGz(archivePath, destDir)
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var total int64
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := validateBundlePath(hdr.Name); err != nil {
			return err
		}
		target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
		if hdr.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		total += hdr.Size
		if total > maxExtractBytes {
			return fmt.Errorf("archive exceeds size limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeExtractedFile(target, tr, hdr.FileInfo().Mode()); err != nil {
			return err
		}
	}
	return nil
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	var total int64
	for _, f := range r.File {
		if err := validateBundlePath(f.Name); err != nil {
			return err
		}
		total += int64(f.UncompressedSize64)
		if total > maxExtractBytes {
			return fmt.Errorf("archive exceeds size limit")
		}
		target := filepath.Join(destDir, filepath.FromSlash(f.Name))
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = writeExtractedFile(target, rc, f.Mode())
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func validateBundlePath(name string) error {
	clean := filepath.ToSlash(filepath.Clean(name))
	if clean == "." || clean == "" {
		return nil
	}
	if strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") || filepath.IsAbs(name) {
		return fmt.Errorf("invalid path %q", name)
	}
	if strings.HasPrefix(clean, "._") || strings.Contains(clean, "/._") || strings.HasPrefix(clean, "__MACOSX") {
		return nil
	}
	parts := strings.Split(strings.TrimPrefix(clean, "./"), "/")
	if len(parts) == 0 {
		return nil
	}
	if len(parts) == 1 {
		if parts[0] == pluginsDirName {
			return nil
		}
		if parts[0] == toolname.AIT || aitplugin.IsOfficialPlugin(parts[0]) || strings.HasSuffix(parts[0], ".exe") {
			return nil
		}
		base := strings.TrimSuffix(parts[0], ".exe")
		if base == toolname.AIT || aitplugin.IsOfficialPlugin(base) {
			return nil
		}
		return fmt.Errorf("unexpected top-level file %q", parts[0])
	}
	if parts[0] != pluginsDirName {
		return fmt.Errorf("unexpected path %q", name)
	}
	if len(parts) == 2 && aitplugin.IsOfficialPlugin(parts[1]) {
		return nil
	}
	if len(parts) != 3 {
		return fmt.Errorf("unexpected plugins path %q", name)
	}
	if !aitplugin.IsOfficialPlugin(parts[1]) {
		return fmt.Errorf("unexpected plugin %q", parts[1])
	}
	if parts[2] != pluginJSONName && parts[2] != parts[1] && parts[2] != parts[1]+".exe" {
		return fmt.Errorf("unexpected plugin file %q", parts[2])
	}
	return nil
}

func writeExtractedFile(target string, r io.Reader, mode fs.FileMode) error {
	perm := mode | 0o755
	if runtime.GOOS == "windows" {
		perm = 0o755
	}
	tmp := target + ".extracting"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if err := os.Rename(tmp, target); err != nil {
		os.Remove(tmp)
		return err
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(target, 0o755)
	}
	return nil
}

func applyBundle(installRoot, extractDir string, plan Plan) error {
	if plan.IncludeAIT {
		if err := installBinary(installRoot, extractDir, aitBinaryName); err != nil {
			return err
		}
	}
	for _, tool := range plan.Tools {
		if err := installBinary(installRoot, extractDir, tool); err != nil {
			return err
		}
		if err := installPluginLayout(installRoot, extractDir, tool); err != nil {
			return err
		}
	}
	return nil
}

func installBinary(installRoot, extractDir, binary string) error {
	src, ok := findBundleBinary(extractDir, binary)
	if !ok {
		return fmt.Errorf("%w: missing binary %q", ErrBundleInvalid, binary)
	}
	for _, name := range shellext.CandidateNames(binary) {
		dest := filepath.Join(installRoot, name)
		if err := atomicCopyFile(src, dest); err != nil {
			return err
		}
		if name == binary || name == binary+".exe" {
			break
		}
	}
	return nil
}

func installPluginLayout(installRoot, extractDir, tool string) error {
	plugDir := filepath.Join(installRoot, pluginsDirName, tool)
	if err := os.MkdirAll(plugDir, 0o755); err != nil {
		return err
	}

	srcJSON := filepath.Join(extractDir, pluginsDirName, tool, pluginJSONName)
	if _, err := os.Stat(srcJSON); err == nil {
		if err := atomicCopyFile(srcJSON, filepath.Join(plugDir, pluginJSONName)); err != nil {
			return err
		}
	} else {
		f, err := os.Create(filepath.Join(plugDir, pluginJSONName))
		if err != nil {
			return err
		}
		if err := aitplugin.WriteManifest(f, tool); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}

	srcBin, ok := findBundleBinary(filepath.Join(extractDir, pluginsDirName, tool), tool)
	if !ok {
		srcBin, ok = findBundleBinary(extractDir, tool)
	}
	if !ok {
		return fmt.Errorf("%w: missing plugin binary %q", ErrBundleInvalid, tool)
	}
	for _, name := range shellext.CandidateNames(tool) {
		dest := filepath.Join(plugDir, name)
		if err := atomicCopyFile(srcBin, dest); err != nil {
			return err
		}
		if name == tool || name == tool+".exe" {
			break
		}
	}
	return nil
}

func findBundleBinary(dir, binary string) (string, bool) {
	for _, name := range shellext.CandidateNames(binary) {
		path := filepath.Join(dir, name)
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		return path, true
	}
	return "", false
}

func atomicCopyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dest + ".new"
	out, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(tmp)
		return err
	}
	if err := out.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(tmp, 0o755)
	}
	if err := os.Rename(tmp, dest); err != nil {
		os.Remove(tmp)
		return err
	}
	return nil
}
