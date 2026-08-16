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

package logsrc

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/os-artificer/AiT/internal/pkg/errhint"
	"github.com/os-artificer/AiT/internal/pkg/toolname"

	"golang.org/x/sys/windows"
)

var (
	modWevtapi    = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtQuery  = modWevtapi.NewProc("EvtQuery")
	procEvtNext   = modWevtapi.NewProc("EvtNext")
	procEvtRender = modWevtapi.NewProc("EvtRender")
	procEvtClose  = modWevtapi.NewProc("EvtClose")
)

const (
	// EvtQuery flags (winevt.h).
	evtQueryChannelPath      = 0x1
	evtQueryReverseDirection = 0x200
	// EvtRender flag selecting the event's XML representation.
	evtRenderEventXML = 1
	// matchAllEvents is the XPath used when no filter narrows the query.
	matchAllEvents = "*"
)

// EvtAPISource reads a Windows Event Log channel in-process through
// wevtapi.dll, so Windows needs no external binary on PATH.
type EvtAPISource struct {
	Channel string
	Label   string
}

// Name returns the channel label.
func (s *EvtAPISource) Name() string {
	if s.Label != "" {
		return s.Label
	}
	return s.Channel
}

// SupportsReverse reports true: last-N queries pass EvtQueryReverseDirection
// so the newest events arrive first.
func (s *EvtAPISource) SupportsReverse() bool { return true }

// Open starts an EvtQuery over the channel, narrowed by an XPath built from
// q when it filters on time or level.
//
// Returns ErrToolMissing when wevtapi.dll cannot be loaded, ErrPermission when
// the channel denies access (the Security channel needs elevation), or
// ErrNoSource when the channel does not exist.
func (s *EvtAPISource) Open(ctx context.Context, q Query) (Cursor, error) {
	if q.Follow {
		return nil, errhint.WithTry(
			ErrFollowUnsupported,
			"Follow is not supported for Windows Event Log.",
			toolname.AILog+" -n 50",
			toolname.AILog+" --since 1h",
		)
	}

	if err := modWevtapi.Load(); err != nil {
		return nil, fmt.Errorf("%w: wevtapi.dll: %v", ErrToolMissing, err)
	}

	channel := s.Channel
	if channel == "" {
		channel = ChannelForSource(q.Source)
	}
	channelPtr, err := windows.UTF16PtrFromString(channel)
	if err != nil {
		return nil, err
	}
	queryPtr, err := windows.UTF16PtrFromString(eventQueryFor(q))
	if err != nil {
		return nil, err
	}

	handle, _, callErr := procEvtQuery.Call(0,
		uintptr(unsafe.Pointer(channelPtr)),
		uintptr(unsafe.Pointer(queryPtr)),
		uintptr(evtQueryFlags(q)))
	if handle == 0 {
		return nil, mapWindowsEvtErr(callErr)
	}
	return &evtAPICursor{ctx: ctx, result: windows.Handle(handle), channel: channel}, nil
}

func evtQueryFlags(q Query) uint32 {
	flags := uint32(evtQueryChannelPath)
	if !q.Head && !q.Follow {
		flags |= evtQueryReverseDirection
	}
	return flags
}

func eventQueryFor(q Query) string {
	if q.Since.IsZero() && q.Until.IsZero() && q.MinLevel == LevelUnset {
		return matchAllEvents
	}
	return buildEventXPath(q)
}
