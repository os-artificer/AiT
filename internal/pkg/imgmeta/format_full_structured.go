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
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	xmlAttrPrefix = "@"
	xmlTextKey    = "#text"
)

type xmlChild struct {
	name  string
	value any
}

// tryExpandStructuredString flattens s when it is structured JSON, XML, or YAML
// (object/array / element tree / mapping-or-sequence). Returns ok=false for plain text.
func tryExpandStructuredString(prefix, s string, depth int) ([]tableRow, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, false
	}

	if rows, ok := expandAsJSON(prefix, s, depth); ok {
		return rows, true
	}
	if rows, ok := expandAsXML(prefix, s, depth); ok {
		return rows, true
	}
	if rows, ok := expandAsYAML(prefix, s, depth); ok {
		return rows, true
	}
	return nil, false
}

func expandAsJSON(prefix, s string, depth int) ([]tableRow, bool) {
	if !jsonLooksLikeObjectOrArray(s) || !json.Valid([]byte(s)) {
		return nil, false
	}
	return rowsFromMarshaled(prefix, json.RawMessage(s), depth)
}

func jsonLooksLikeObjectOrArray(s string) bool {
	return s != "" && (s[0] == '{' || s[0] == '[')
}

func expandAsXML(prefix, s string, depth int) ([]tableRow, bool) {
	if !strings.HasPrefix(s, "<") {
		return nil, false
	}

	root, err := parseXMLDocument(s)
	if err != nil || root == nil {
		return nil, false
	}
	return rowsFromValue(prefix, root, depth)
}

func expandAsYAML(prefix, s string, depth int) ([]tableRow, bool) {
	// XML leftovers are handled by expandAsXML; skip here.
	if strings.HasPrefix(s, "<") {
		return nil, false
	}

	var v any
	if err := yaml.Unmarshal([]byte(s), &v); err != nil {
		return nil, false
	}
	switch v.(type) {
	case map[string]any, []any:
		// structured only
	default:
		return nil, false
	}
	return rowsFromValue(prefix, v, depth)
}

func rowsFromValue(prefix string, v any, depth int) ([]tableRow, bool) {
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, false
	}
	return rowsFromMarshaled(prefix, raw, depth)
}

func rowsFromMarshaled(prefix string, raw json.RawMessage, depth int) ([]tableRow, bool) {
	var rows []tableRow
	appendValueRows(&rows, prefix, raw, depth)
	if len(rows) == 0 {
		return nil, false
	}
	return rows, true
}

func parseXMLDocument(s string) (any, error) {
	dec := xml.NewDecoder(strings.NewReader(s))
	dec.Strict = false

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil, fmt.Errorf("xml: no root element")
		}
		if err != nil {
			return nil, err
		}
		se, ok := tok.(xml.StartElement)
		if !ok {
			continue
		}

		node, err := parseXMLElement(dec, se)
		if err != nil {
			return nil, err
		}
		return map[string]any{se.Name.Local: node}, nil
	}
}

func parseXMLElement(dec *xml.Decoder, start xml.StartElement) (any, error) {
	attrs := make(map[string]any, len(start.Attr))
	for _, a := range start.Attr {
		attrs[xmlAttrPrefix+a.Name.Local] = a.Value
	}

	var (
		text     strings.Builder
		children []xmlChild
	)

	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child, err := parseXMLElement(dec, t)
			if err != nil {
				return nil, err
			}
			children = append(children, xmlChild{name: t.Name.Local, value: child})
		case xml.EndElement:
			return finalizeXMLNode(attrs, text.String(), children), nil
		case xml.CharData:
			appendXMLText(&text, t)
		}
	}
}

func appendXMLText(text *strings.Builder, data xml.CharData) {
	chunk := bytes.TrimSpace(data)
	if len(chunk) == 0 {
		return
	}
	if text.Len() > 0 {
		text.WriteByte(' ')
	}
	text.Write(chunk)
}

func finalizeXMLNode(attrs map[string]any, text string, children []xmlChild) any {
	if len(children) == 0 && len(attrs) == 0 {
		return text
	}

	out := make(map[string]any, len(attrs)+len(children)+1)
	for k, v := range attrs {
		out[k] = v
	}
	if text != "" {
		out[xmlTextKey] = text
	}
	mergeXMLChildren(out, children)
	return out
}

func mergeXMLChildren(out map[string]any, children []xmlChild) {
	counts := make(map[string]int, len(children))
	for _, c := range children {
		counts[c.name]++
	}
	for _, c := range children {
		if counts[c.name] == 1 {
			out[c.name] = c.value
			continue
		}
		arr, _ := out[c.name].([]any)
		out[c.name] = append(arr, c.value)
	}
}
