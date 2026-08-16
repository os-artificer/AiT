//go:build windows

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

package hoststats

import (
	"context"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modIPHlpAPI            = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetTcpStatisticsEx = modIPHlpAPI.NewProc("GetTcpStatisticsEx")
	procGetUdpStatisticsEx = modIPHlpAPI.NewProc("GetUdpStatisticsEx")
	procGetIcmpStatistics  = modIPHlpAPI.NewProc("GetIcmpStatistics")
	procGetIpNetTable      = modIPHlpAPI.NewProc("GetIpNetTable")
)

const (
	winAFInet            = 2
	winAFInet6           = 23
	winMIBNetTypeInvalid = 2
	winIPNetTableHdr     = 4
)

type mibTCPStats struct {
	RtoAlgorithm uint32
	RtoMin       uint32
	RtoMax       uint32
	MaxConn      uint32
	ActiveOpens  uint32
	PassiveOpens uint32
	AttemptFails uint32
	EstabResets  uint32
	CurrEstab    uint32
	InSegs       uint32
	OutSegs      uint32
	RetransSegs  uint32
	InErrs       uint32
	OutRsts      uint32
	NumConns     uint32
}

type mibUDPStats struct {
	InDatagrams  uint32
	NoPorts      uint32
	InErrors     uint32
	OutDatagrams uint32
	NumAddrs     uint32
}

type mibICMPStats struct {
	Msgs          uint32
	Errors        uint32
	DestUnreachs  uint32
	TimeExcds     uint32
	ParmProbs     uint32
	SrcQuenchs    uint32
	Redirects     uint32
	Echos         uint32
	EchoReps      uint32
	Timestamps    uint32
	TimestampReps uint32
	AddrMasks     uint32
	AddrMaskReps  uint32
}

type mibICMPInfo struct {
	InStats  mibICMPStats
	OutStats mibICMPStats
}

type mibIPNetRow struct {
	Index       uint32
	PhysAddrLen uint32
	PhysAddr    [8]byte
	Addr        uint32
	Type        uint32
}

func sampleProtoRaw(ctx context.Context) (protoRaw, error) {
	_ = ctx
	var raw protoRaw
	fillWindowsTCP(&raw)
	fillWindowsUDP(&raw)
	fillWindowsICMP(&raw)
	fillWindowsARP(&raw)
	return raw, nil
}

func fillWindowsTCP(raw *protoRaw) {
	mergeWindowsTCP(raw, winAFInet)
	mergeWindowsTCP(raw, winAFInet6)
}

func mergeWindowsTCP(raw *protoRaw, family uintptr) {
	var st mibTCPStats
	r, _, _ := procGetTcpStatisticsEx.Call(uintptr(unsafe.Pointer(&st)), family)
	if r != 0 {
		return
	}
	raw.tcpIn += uint64(st.InSegs)
	raw.tcpOut += uint64(st.OutSegs)
	raw.tcpRetrans += uint64(st.RetransSegs)
	raw.tcpEstab += uint64(st.CurrEstab)
	raw.tcpOpens += uint64(st.ActiveOpens) + uint64(st.PassiveOpens)
}

func fillWindowsUDP(raw *protoRaw) {
	mergeWindowsUDP(raw, winAFInet)
	mergeWindowsUDP(raw, winAFInet6)
}

func mergeWindowsUDP(raw *protoRaw, family uintptr) {
	var st mibUDPStats
	r, _, _ := procGetUdpStatisticsEx.Call(uintptr(unsafe.Pointer(&st)), family)
	if r != 0 {
		return
	}
	raw.udpIn += uint64(st.InDatagrams)
	raw.udpOut += uint64(st.OutDatagrams)
	raw.udpNoPort += uint64(st.NoPorts)
	raw.udpInErr += uint64(st.InErrors)
}

func fillWindowsICMP(raw *protoRaw) {
	var st mibICMPInfo
	r, _, _ := procGetIcmpStatistics.Call(uintptr(unsafe.Pointer(&st)))
	if r != 0 {
		return
	}
	raw.icmpIn = uint64(st.InStats.Msgs)
	raw.icmpOut = uint64(st.OutStats.Msgs)
	raw.icmpInErr = uint64(st.InStats.Errors)
	raw.icmpOutErr = uint64(st.OutStats.Errors)
}

func fillWindowsARP(raw *protoRaw) {
	var size uint32
	r, _, _ := procGetIpNetTable.Call(0, uintptr(unsafe.Pointer(&size)), 0)
	if size == 0 {
		return
	}
	buf := make([]byte, size)
	r, _, _ = procGetIpNetTable.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0)
	if r != 0 {
		return
	}
	appendWindowsARP(raw, buf)
}

func appendWindowsARP(raw *protoRaw, buf []byte) {
	if len(buf) < winIPNetTableHdr {
		return
	}
	count := *(*uint32)(unsafe.Pointer(&buf[0]))
	rowSize := int(unsafe.Sizeof(mibIPNetRow{}))
	off := winIPNetTableHdr
	for i := uint32(0); i < count; i++ {
		if off+rowSize > len(buf) {
			return
		}
		row := (*mibIPNetRow)(unsafe.Pointer(&buf[off]))
		raw.arpEnts++
		if row.Type == winMIBNetTypeInvalid || row.PhysAddrLen == 0 {
			raw.arpIncomp++
		}
		off += rowSize
	}
}
