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
	"encoding/json"
	"io"
)

type jsonGPS struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type jsonObject struct {
	File         string          `json:"file"`
	Format       string          `json:"format"`
	Width        int             `json:"width"`
	Height       int             `json:"height"`
	Make         string          `json:"make"`
	Model        string          `json:"model"`
	Software     string          `json:"software"`
	LensMake     string          `json:"lens_make"`
	LensModel    string          `json:"lens_model"`
	SerialNumber string          `json:"serial_number"`
	DateTime     string          `json:"datetime"`
	Address      string          `json:"address"`
	GPS          jsonGPS         `json:"gps"`
	Altitude     float64         `json:"altitude"`
	Exif         json.RawMessage `json:"exif,omitempty"`
}

// FormatJSON writes a JSON array of info objects to w (always an array).
// Locked keys are always present; missing string-like values are "0".
// When full is true, each object includes an "exif" key from the library marshal of Info.Exif.
func FormatJSON(w io.Writer, infos []Info, full bool) error {
	objs := make([]jsonObject, 0, len(infos))
	for _, info := range infos {
		o, err := toJSONObject(info, full)
		if err != nil {
			return err
		}
		objs = append(objs, o)
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(objs)
}

func toJSONObject(info Info, full bool) (jsonObject, error) {
	lat, lon := gpsOrZero(info)
	o := jsonObject{
		File:         displayOrZero(info.File),
		Format:       displayOrZero(info.Format),
		Width:        info.Width,
		Height:       info.Height,
		Make:         displayOrZero(info.Make),
		Model:        displayOrZero(info.Model),
		Software:     displayOrZero(info.Software),
		LensMake:     displayOrZero(info.LensMake),
		LensModel:    displayOrZero(info.LensModel),
		SerialNumber: displayOrZero(info.SerialNumber),
		DateTime:     dateTimeOrZero(info.DateTime),
		Address:      displayOrZero(info.Address),
		GPS:          jsonGPS{Lat: lat, Lon: lon},
		Altitude:     altitudeOrZero(info),
	}
	if !full {
		return o, nil
	}
	raw, err := MarshalFullExif(info.Exif)
	if err != nil {
		return jsonObject{}, err
	}
	o.Exif = raw
	return o, nil
}
