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

package hostdash

import (
	"fmt"
	"time"

	"github.com/os-artificer/AiT/internal/pkg/hoststats"
)

const (
	protoTCP protoKind = iota
	protoUDP
	protoARP
	protoICMP
)

type protoKind int

// ProtoPanel shows one protocol (TCP/UDP/ARP/ICMP) with NET-style rates and features.
type ProtoPanel struct {
	kind    protoKind
	title   string
	histIn  []float64
	histOut []float64
	times   []time.Time
	value   string
	styled  string
	detail  string
}

// NewTCPPanel returns the TCP protocol panel.
func NewTCPPanel() *ProtoPanel {
	return newProtoPanel(protoTCP, titleTCP)
}

// NewUDPPanel returns the UDP protocol panel.
func NewUDPPanel() *ProtoPanel {
	return newProtoPanel(protoUDP, titleUDP)
}

// NewARPPanel returns the ARP protocol panel.
func NewARPPanel() *ProtoPanel {
	return newProtoPanel(protoARP, titleARP)
}

// NewICMPPanel returns the ICMP protocol panel.
func NewICMPPanel() *ProtoPanel {
	return newProtoPanel(protoICMP, titleICMP)
}

func newProtoPanel(kind protoKind, title string) *ProtoPanel {
	return &ProtoPanel{
		kind:    kind,
		title:   title,
		histIn:  make([]float64, 0, sparkLen),
		histOut: make([]float64, 0, sparkLen),
		times:   make([]time.Time, 0, sparkLen),
	}
}

// Title implements Panel.
func (p *ProtoPanel) Title() string { return p.title }

// Update implements Panel.
func (p *ProtoPanel) Update(s hoststats.Snapshot) {
	var in, out float64
	switch p.kind {
	case protoTCP:
		t := s.Proto.TCP
		in, out = t.InPerSec, t.OutPerSec
		p.detail = fmt.Sprintf("estab %d  retrans %s  opens %d",
			t.Estab, hoststats.FormatPacketRate(t.RetransPerSec)+rateUnit, t.Opens)
	case protoUDP:
		u := s.Proto.UDP
		in, out = u.InPerSec, u.OutPerSec
		p.detail = fmt.Sprintf("noport %d  inerr %s",
			u.NoPort, colorSparkUint(u.InErr, true))
	case protoARP:
		a := s.Proto.ARP
		in, out = a.InPerSec, a.OutPerSec
		p.detail = fmt.Sprintf("ents %d  incomp %d  lookup %s",
			a.Entries, a.Incomplete, hoststats.FormatPacketRate(a.LookupsPerSec)+rateUnit)
	default:
		i := s.Proto.ICMP
		in, out = i.InPerSec, i.OutPerSec
		p.detail = fmt.Sprintf("inerr %s  outerr %s",
			colorSparkUint(i.InErr, true),
			colorSparkUint(i.OutErr, false))
	}

	p.value, p.styled = formatInOutHeadline(
		hoststats.FormatPacketRate(in)+rateUnit,
		hoststats.FormatPacketRate(out)+rateUnit)

	p.histIn, p.histOut, p.times = appendSparkDual(p.histIn, p.histOut, p.times, in, out, s.Time)
}

// View implements Panel. width/height are outer terminal cells including border.
func (p *ProtoPanel) View(width, height int) string {
	return renderDualSparkView(
		p.title, p.value, p.styled, p.detail,
		p.histIn, p.histOut, p.times, width, height)
}
