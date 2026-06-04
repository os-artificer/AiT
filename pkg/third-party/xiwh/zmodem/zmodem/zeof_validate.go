package zmodem

import (
	"encoding/binary"

	"github.com/os-artificer/AiT/pkg/third-party/xiwh/zmodem/byteutil"
)

// zeofOffsetMatches reports whether the ZEOF header offset equals bytes read from buf.
// When headerData is shorter than 4 bytes, offset validation is skipped (legacy peers).
func zeofOffsetMatches(headerData []byte, buf *byteutil.BlockBuffer) bool {
	if buf == nil || len(headerData) < 4 {
		return true
	}
	offset := int64(binary.LittleEndian.Uint32(headerData))
	return offset == buf.BytesReceived()
}
