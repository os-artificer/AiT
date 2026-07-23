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

package imgmeta

import (
	"image"
	"os"
	"path/filepath"
	"strings"

	"github.com/evanoberholster/imagemeta/imagetype"
	"github.com/evanoberholster/imagemeta/meta/exif"
)

func attachExif(info *Info, ex exif.Exif) {
	cp := ex
	info.Exif = &cp
}

// fillDimsFromFile reads image dimensions from an open file via Seek + DecodeConfig.
func fillDimsFromFile(info *Info, f *os.File) {
	if f == nil {
		return
	}
	if _, err := f.Seek(0, 0); err != nil {
		return
	}
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return
	}
	if info.Width == 0 {
		info.Width = cfg.Width
	}
	if info.Height == 0 {
		info.Height = cfg.Height
	}
}

func formatFromExt(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "jpg", "jpeg", "jpe", "jfif":
		return "jpeg"
	case "png":
		return "png"
	case "tif", "tiff":
		return "tiff"
	case "webp":
		return "webp"
	case "heic", "heif":
		return "heic"
	case "exif":
		return "exif"
	default:
		return ext
	}
}

func formatFromImageType(it imagetype.ImageType) string {
	if it.IsUnknown() {
		return ""
	}
	ext := it.Extension()
	switch ext {
	case "jpg", "jpeg":
		return "jpeg"
	default:
		return ext
	}
}
