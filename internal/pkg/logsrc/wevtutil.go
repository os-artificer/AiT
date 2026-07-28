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
	"strconv"
	"strings"
	"time"
)

// BuildWevtutilArgs builds argv for wevtutil qe (without the binary name).
func BuildWevtutilArgs(channel string, q Query, reverse bool) []string {
	args := []string{"qe", channel, "/f:xml", "/rd:true"}
	if !reverse {
		args[len(args)-1] = "/rd:false"
	}
	if !q.Since.IsZero() || !q.Until.IsZero() || q.MinLevel != LevelUnset {
		args = append(args, "/q:"+buildEventXPath(q))
	}
	return args
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
		Computer string `xml:"Computer"`
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

// ParseEventXML parses one Windows Event XML document into a Record.
func ParseEventXML(data []byte) (Record, error) {
	var ev wevtEvent
	if err := xml.Unmarshal(data, &ev); err != nil {
		return Record{}, err
	}
	msg := strings.TrimSpace(ev.RenderingInfo.Message)
	if msg == "" {
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
		msg = b.String()
	}
	rec := Record{
		Raw:     string(data),
		Host:    ev.System.Computer,
		Unit:    ev.System.Provider.Name,
		Message: msg,
		Level:   FromWindowsLevel(ev.System.Level),
		Source:  "system",
	}
	if ts := ev.System.TimeCreated.SystemTime; ts != "" {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			rec.Time = t
		} else if t, err := time.Parse(time.RFC3339, ts); err == nil {
			rec.Time = t
		}
	}
	return rec, nil
}

// ChannelForSource maps logical source names to Windows Event Log channels.
func ChannelForSource(name string) string {
	switch strings.ToLower(name) {
	case "application":
		return "Application"
	case "security":
		return "Security"
	case "setup":
		return "Setup"
	default:
		return "System"
	}
}
