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
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/dsoprea/go-exif/v3"
	exifcommon "github.com/dsoprea/go-exif/v3/common"
	jpegstructure "github.com/dsoprea/go-jpeg-image-structure/v2"
)

const (
	ifdPathRoot = "IFD0"
	ifdPathExif = "IFD0/Exif"
	ifdPathGPS  = "IFD0/GPSInfo"
)

// Patch writes the non-nil fields in p into path (JPEG only, in place).
// Returns ErrEmptyPath, ErrNothingToChange, ErrUnsupported, or a write/parse error.
func Patch(path string, p PatchFields) error {
	return PatchTo(path, path, p)
}

// PatchTo writes the non-nil fields in p by reading src and writing dst (JPEG only).
// src and dst must be non-empty. When dst differs from src, src is left unchanged.
// Returns ErrEmptyPath, ErrNothingToChange, ErrUnsupported, or a write/parse error.
func PatchTo(src, dst string, p PatchFields) error {
	if src == "" || dst == "" {
		return ErrEmptyPath
	}
	if p.Empty() {
		return ErrNothingToChange
	}
	if p.Latitude != nil || p.Longitude != nil {
		if !p.HasGPS() {
			return fmt.Errorf("%w: --gps requires both latitude and longitude", ErrNothingToChange)
		}
	}
	return mutateJPEG(src, dst, func(rootIb *exif.IfdBuilder) error {
		return applyPatch(rootIb, p)
	})
}

// Clear removes the selected writable EXIF fields from path (JPEG only, in place).
// Returns ErrEmptyPath, ErrNothingToChange, ErrUnsupported, or a write/parse error.
func Clear(path string, mask ClearMask) error {
	return ClearTo(path, path, mask)
}

// ClearTo removes the selected writable EXIF fields by reading src and writing dst (JPEG only).
// src and dst must be non-empty. When dst differs from src, src is left unchanged.
// Returns ErrEmptyPath, ErrNothingToChange, ErrUnsupported, or a write/parse error.
func ClearTo(src, dst string, mask ClearMask) error {
	if src == "" || dst == "" {
		return ErrEmptyPath
	}
	if mask.Empty() {
		return ErrNothingToChange
	}
	return mutateJPEG(src, dst, func(rootIb *exif.IfdBuilder) error {
		return applyClear(rootIb, mask)
	})
}

func mutateJPEG(src, dst string, apply func(*exif.IfdBuilder) error) error {
	if err := requireJPEG(src); err != nil {
		return err
	}

	var srcMode os.FileMode
	if fi, err := os.Stat(src); err == nil {
		srcMode = fi.Mode().Perm()
	}

	jmp := jpegstructure.NewJpegMediaParser()
	intfc, err := jmp.ParseFile(src)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnsupported, err)
	}
	sl, ok := intfc.(*jpegstructure.SegmentList)
	if !ok {
		return ErrUnsupported
	}

	rootIb, err := sl.ConstructExifBuilder()
	if err != nil {
		return err
	}
	if err := apply(rootIb); err != nil {
		return err
	}
	if err := sl.SetExif(rootIb); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(dst), ".aiimg-*.jpg")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if err := sl.Write(tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := replaceFile(tmpName, dst); err != nil {
		return err
	}
	if srcMode != 0 {
		_ = os.Chmod(dst, srcMode)
	}
	return nil
}

// replaceFile moves tmp onto dst. On Windows, removes dst first because Rename fails if dest exists.
func replaceFile(tmp, dst string) error {
	if runtime.GOOS == "windows" {
		_ = os.Remove(dst)
	}
	return os.Rename(tmp, dst)
}

func requireJPEG(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var hdr [3]byte
	n, err := f.Read(hdr[:])
	if err != nil || n < 3 {
		return ErrUnsupported
	}
	if hdr[0] != 0xff || hdr[1] != 0xd8 || hdr[2] != 0xff {
		return ErrUnsupported
	}
	return nil
}

func applyPatch(rootIb *exif.IfdBuilder, p PatchFields) error {
	ifd0, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathRoot)
	if err != nil {
		return err
	}

	if err := setASCII(ifd0, "Make", p.Make); err != nil {
		return err
	}
	if err := setASCII(ifd0, "Model", p.Model); err != nil {
		return err
	}
	if err := setASCII(ifd0, "Software", p.Software); err != nil {
		return err
	}
	if p.DateTime != nil {
		if err := ifd0.SetStandardWithName("DateTime", *p.DateTime); err != nil {
			return err
		}
	}

	needExif := p.LensMake != nil || p.LensModel != nil || p.SerialNumber != nil || p.DateTime != nil
	if needExif {
		exifIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathExif)
		if err != nil {
			return err
		}
		if err := setASCII(exifIb, "LensMake", p.LensMake); err != nil {
			return err
		}
		if err := setASCII(exifIb, "LensModel", p.LensModel); err != nil {
			return err
		}
		if err := setASCII(exifIb, "BodySerialNumber", p.SerialNumber); err != nil {
			return err
		}
		if p.DateTime != nil {
			if err := exifIb.SetStandardWithName("DateTimeOriginal", *p.DateTime); err != nil {
				return err
			}
		}
	}

	if p.HasGPS() || p.Altitude != nil {
		gpsIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathGPS)
		if err != nil {
			return err
		}
		if p.HasGPS() {
			if err := setGPSCoords(gpsIb, *p.Latitude, *p.Longitude); err != nil {
				return err
			}
		}
		if p.Altitude != nil {
			if err := setGPSAltitude(gpsIb, *p.Altitude); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyClear(rootIb *exif.IfdBuilder, mask ClearMask) error {
	ifd0, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathRoot)
	if err != nil {
		return err
	}
	if mask.Make {
		_ = deleteTagByName(ifd0, "Make")
	}
	if mask.Model {
		_ = deleteTagByName(ifd0, "Model")
	}
	if mask.Software {
		_ = deleteTagByName(ifd0, "Software")
	}
	if mask.DateTime {
		_ = deleteTagByName(ifd0, "DateTime")
	}

	needExif := mask.LensMake || mask.LensModel || mask.SerialNumber || mask.DateTime
	if needExif {
		exifIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathExif)
		if err != nil {
			return err
		}
		if mask.LensMake {
			_ = deleteTagByName(exifIb, "LensMake")
		}
		if mask.LensModel {
			_ = deleteTagByName(exifIb, "LensModel")
		}
		if mask.SerialNumber {
			_ = deleteTagByName(exifIb, "BodySerialNumber")
		}
		if mask.DateTime {
			_ = deleteTagByName(exifIb, "DateTimeOriginal")
		}
	}

	if mask.GPS || mask.Altitude {
		gpsIb, err := exif.GetOrCreateIbFromRootIb(rootIb, ifdPathGPS)
		if err != nil {
			return err
		}
		if mask.GPS {
			for _, name := range []string{"GPSLatitude", "GPSLatitudeRef", "GPSLongitude", "GPSLongitudeRef"} {
				_ = deleteTagByName(gpsIb, name)
			}
		}
		if mask.Altitude {
			for _, name := range []string{"GPSAltitude", "GPSAltitudeRef"} {
				_ = deleteTagByName(gpsIb, name)
			}
		}
	}
	return nil
}

func setASCII(ib *exif.IfdBuilder, name string, v *string) error {
	if v == nil {
		return nil
	}
	return ib.SetStandardWithName(name, *v)
}

var clearTagIDs = map[string]uint16{
	"Make":             0x010f,
	"Model":            0x0110,
	"Software":         0x0131,
	"DateTime":         0x0132,
	"DateTimeOriginal": 0x9003,
	"BodySerialNumber": 0xa431,
	"LensMake":         0xa433,
	"LensModel":        0xa434,
	"GPSLatitudeRef":   0x0001,
	"GPSLatitude":      0x0002,
	"GPSLongitudeRef":  0x0003,
	"GPSLongitude":     0x0004,
	"GPSAltitudeRef":   0x0005,
	"GPSAltitude":      0x0006,
}

func deleteTagByName(ib *exif.IfdBuilder, name string) error {
	id, ok := clearTagIDs[name]
	if !ok {
		return nil
	}
	_ = ib.DeleteFirst(id)
	return nil
}

func setGPSCoords(gpsIb *exif.IfdBuilder, lat, lon float64) error {
	latDeg := decimalToGpsDegrees(lat, 'S', 'N')
	lonDeg := decimalToGpsDegrees(lon, 'W', 'E')

	if err := gpsIb.SetStandardWithName("GPSLatitudeRef", string([]byte{latDeg.Orientation})); err != nil {
		return err
	}
	if err := gpsIb.SetStandardWithName("GPSLatitude", gpsDegreesRationals(latDeg)); err != nil {
		return err
	}
	if err := gpsIb.SetStandardWithName("GPSLongitudeRef", string([]byte{lonDeg.Orientation})); err != nil {
		return err
	}
	return gpsIb.SetStandardWithName("GPSLongitude", gpsDegreesRationals(lonDeg))
}

func setGPSAltitude(gpsIb *exif.IfdBuilder, alt float64) error {
	ref := uint8(0)
	if alt < 0 {
		ref = 1
		alt = -alt
	}
	if err := gpsIb.SetStandardWithName("GPSAltitudeRef", []byte{ref}); err != nil {
		return err
	}
	raw := []exifcommon.Rational{{Numerator: uint32(math.Round(alt * 100)), Denominator: 100}}
	return gpsIb.SetStandardWithName("GPSAltitude", raw)
}

func decimalToGpsDegrees(v float64, negRef, posRef byte) exif.GpsDegrees {
	ref := posRef
	if v < 0 {
		ref = negRef
		v = -v
	}
	deg := math.Floor(v)
	minFloat := (v - deg) * 60
	min := math.Floor(minFloat)
	sec := (minFloat - min) * 60
	return exif.GpsDegrees{Orientation: ref, Degrees: deg, Minutes: min, Seconds: sec}
}

func gpsDegreesRationals(d exif.GpsDegrees) []exifcommon.Rational {
	return []exifcommon.Rational{
		{Numerator: uint32(d.Degrees), Denominator: 1},
		{Numerator: uint32(d.Minutes), Denominator: 1},
		{Numerator: uint32(math.Round(d.Seconds * 100)), Denominator: 100},
	}
}

// ParseDateTime parses RFC3339 or EXIF "2006:01:02 15:04:05" into UTC wall time.
func ParseDateTime(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t.UTC(), nil
	}
	t, err := time.ParseInLocation("2006:01:02 15:04:05", s, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid datetime %q: want RFC3339 or YYYY:MM:DD HH:MM:SS", s)
	}
	return t, nil
}
