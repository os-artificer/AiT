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
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
	"time"

	"github.com/evanoberholster/imagemeta"
	"github.com/evanoberholster/imagemeta/meta/exif"
)

// Extract opens path and returns best-effort metadata.
// Missing EXIF is not an error. Hard-fail: empty path, not found/unreadable, unrecognized format.
func Extract(path string) (Info, error) {
	if strings.TrimSpace(path) == "" {
		return Info{}, ErrEmptyPath
	}

	f, err := os.Open(path)
	if err != nil {
		return Info{}, err
	}
	defer f.Close()

	info := Info{File: path}

	ex, decodeErr := imagemeta.Decode(f)
	if decodeErr == nil {
		fillFromDecoded(&info, ex, f)
		return info, nil
	}

	if errors.Is(decodeErr, imagemeta.ErrImageTypeNotFound) {
		return tryImageConfigFallback(f, path, info, decodeErr)
	}

	if errors.Is(decodeErr, imagemeta.ErrNoExif) {
		attachExif(&info, ex)
		info.Format = formatFromImageType(ex.ImageType)
		if info.Format == "" {
			info.Format = formatFromExt(path)
		}
		fillDimsFromFile(&info, f)
		return info, nil
	}

	// Partial decode with known type: keep available tags.
	if !ex.ImageType.IsUnknown() {
		fillFromDecoded(&info, ex, f)
		return info, nil
	}

	return tryImageConfigFallback(f, path, info, decodeErr)
}

func tryImageConfigFallback(f *os.File, path string, info Info, decodeErr error) (Info, error) {
	if _, seekErr := f.Seek(0, 0); seekErr != nil {
		return Info{}, fmt.Errorf("%w: %v", ErrUnsupported, decodeErr)
	}

	cfg, formatName, cfgErr := image.DecodeConfig(f)
	if cfgErr != nil {
		return Info{}, fmt.Errorf("%w: %v", ErrUnsupported, decodeErr)
	}

	info.Format = formatFromExt(path)
	if info.Format == "" {
		info.Format = strings.ToLower(formatName)
	}
	if info.Format == "" {
		info.Format = "image"
	}
	info.Width = cfg.Width
	info.Height = cfg.Height
	return info, nil
}

func fillFromDecoded(info *Info, ex exif.Exif, f *os.File) {
	attachExif(info, ex)

	info.Format = formatFromImageType(ex.ImageType)
	if info.Format == "" {
		info.Format = formatFromExt(info.File)
	}

	info.Make = strings.TrimSpace(ex.IFD0.Make)
	info.Model = strings.TrimSpace(ex.IFD0.Model)
	info.Software = strings.TrimSpace(ex.IFD0.Software)
	info.LensMake = strings.TrimSpace(ex.ExifIFD.LensMake)
	info.LensModel = strings.TrimSpace(ex.ExifIFD.LensModel)

	serial := strings.TrimSpace(ex.ExifIFD.BodySerialNumber)
	if serial == "" {
		serial = strings.TrimSpace(ex.CameraSerial)
	}
	info.SerialNumber = serial

	if w := int(ex.ExifIFD.PixelXDimension); w > 0 {
		info.Width = w
	} else if w := int(ex.IFD0.ImageWidth); w > 0 {
		info.Width = w
	}
	if h := int(ex.ExifIFD.PixelYDimension); h > 0 {
		info.Height = h
	} else if h := int(ex.IFD0.ImageHeight); h > 0 {
		info.Height = h
	}

	dto := ex.ExifIFD.DateTimeOriginal
	if dto.IsZero() {
		dto = ex.IFD0.ModifyDate
	}
	if !dto.IsZero() {
		// Wall time without timezone: rebuild as local components in UTC clock face.
		info.DateTime = time.Date(dto.Year(), dto.Month(), dto.Day(), dto.Hour(), dto.Minute(), dto.Second(), 0, time.UTC)
	}

	lat := ex.GPS.Latitude()
	lon := ex.GPS.Longitude()
	if lat != 0 || lon != 0 {
		latCopy, lonCopy := lat, lon
		info.Latitude = &latCopy
		info.Longitude = &lonCopy
	}

	alt := float64(ex.GPS.Altitude())
	if alt != 0 {
		info.Altitude = &alt
	}

	if info.Width == 0 || info.Height == 0 {
		fillDimsFromFile(info, f)
	}
}
