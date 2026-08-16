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

package logsrc

import (
	"encoding/xml"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	wevtQueryFmt     = "qe"
	wevtFormatXML    = "/f:xml"
	wevtReverseTrue  = "/rd:true"
	wevtReverseFalse = "/rd:false"

	winEvtProcessName = "processname"
	winEvtImageName   = "imagename"
	winEvtApplication = "application"

	winChannelApplication = "Application"
	winChannelSecurity    = "Security"
	winChannelSetup       = "Setup"
	winChannelSystem      = "System"
)

type wevtEvent struct {
	XMLName xml.Name `xml:"Event"`
	System  struct {
		Provider struct {
			Name string `xml:"Name,attr"`
		} `xml:"Provider"`
		EventID     int `xml:"EventID"`
		Level       int `xml:"Level"`
		TimeCreated struct {
			SystemTime string `xml:"SystemTime,attr"`
		} `xml:"TimeCreated"`
		Computer  string `xml:"Computer"`
		Execution struct {
			ProcessID uint32 `xml:"ProcessID,attr"`
		} `xml:"Execution"`
	} `xml:"System"`
	EventData struct {
		Data []struct {
			Name  string `xml:"Name,attr"`
			Value string `xml:",chardata"`
		} `xml:"Data"`
	} `xml:"EventData"`
	RenderingInfo struct {
		Message string `xml:"Message"`
	} `xml:"RenderingInfo"`
}

// BuildWevtutilArgs builds argv for wevtutil qe (without the binary name).
func BuildWevtutilArgs(channel string, q Query, reverse bool) []string {
	rd := wevtReverseTrue
	if !reverse {
		rd = wevtReverseFalse
	}
	args := []string{wevtQueryFmt, channel, wevtFormatXML, rd}
	if !q.Since.IsZero() || !q.Until.IsZero() || q.MinLevel != LevelUnset {
		args = append(args, "/q:"+buildEventXPath(q))
	}
	return args
}

// ParseEventXML parses one Windows Event XML document into a Record.
func ParseEventXML(data []byte) (Record, error) {
	var ev wevtEvent
	if err := xml.Unmarshal(data, &ev); err != nil {
		return Record{}, err
	}
	rec := Record{
		Raw:     string(data),
		Host:    ev.System.Computer,
		Unit:    ev.System.Provider.Name,
		Message: eventDataMessage(ev),
		Level:   FromWindowsLevel(ev.System.Level),
		Source:  SourceNameSystem,
		PID:     int(ev.System.Execution.ProcessID),
		Process: windowsEventProcessName(ev),
		Time:    parseEventSystemTime(ev.System.TimeCreated.SystemTime),
	}
	return rec, nil
}

// ChannelForSource maps logical source names to Windows Event Log channels.
func ChannelForSource(name string) string {
	switch strings.ToLower(name) {
	case "application":
		return winChannelApplication
	case "security":
		return winChannelSecurity
	case "setup":
		return winChannelSetup
	default:
		return winChannelSystem
	}
}

func eventDataMessage(ev wevtEvent) string {
	if msg := strings.TrimSpace(ev.RenderingInfo.Message); msg != "" {
		return msg
	}
	var b strings.Builder
	for i, d := range ev.EventData.Data {
		if i > 0 {
			b.WriteByte(' ')
		}
		if d.Name != "" {
			b.WriteString(d.Name)
			b.WriteByte('=')
		}
		b.WriteString(d.Value)
	}
	return b.String()
}

func parseEventSystemTime(ts string) time.Time {
	if ts == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t
	}
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return time.Time{}
	}
	return t
}

func buildEventXPath(q Query) string {
	var parts []string
	if q.MinLevel != LevelUnset {
		parts = append(parts, "Level<="+strconv.Itoa(syslogToWindowsLevel(q.MinLevel)))
	}
	if !q.Since.IsZero() {
		parts = append(parts, "TimeCreated[@SystemTime>='"+q.Since.UTC().Format(time.RFC3339)+"']")
	}
	if !q.Until.IsZero() {
		parts = append(parts, "TimeCreated[@SystemTime<='"+q.Until.UTC().Format(time.RFC3339)+"']")
	}
	if len(parts) == 0 {
		return "*"
	}
	return "*[System[" + strings.Join(parts, " and ") + "]]"
}

func syslogToWindowsLevel(lv Level) int {
	switch {
	case lv <= LevelCritical:
		return 1
	case lv <= LevelError:
		return 2
	case lv <= LevelWarning:
		return 3
	case lv <= LevelInfo:
		return 4
	default:
		return 5
	}
}

func windowsEventProcessName(ev wevtEvent) string {
	for _, d := range ev.EventData.Data {
		if d.Value == "" {
			continue
		}
		switch strings.ToLower(d.Name) {
		case winEvtProcessName, winEvtImageName, winEvtApplication:
			return windowsBaseName(d.Value)
		}
	}
	return ""
}

func windowsBaseName(p string) string {
	return path.Base(strings.ReplaceAll(strings.TrimSpace(p), "\\", "/"))
}
