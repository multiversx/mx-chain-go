package slash

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

// HeaderInfo contains a HeaderHandler and its associated hash
type HeaderInfo struct {
	Header data.HeaderHandler
	Hash   []byte
}

// HeaderInfoList defines a list of HeaderInfo
type HeaderInfoList []*HeaderInfo

// SlashingResult contains the slashable data as well as the severity(slashing level)
// for a possible malicious validator
type SlashingResult struct {
	SlashingLevel ThreatLevel
	Headers       HeaderInfoList
}

// IsIndexSetInBitmap - checks if a bit is set(1) in the given bitmap
// TODO: Move this utility function in ELROND-GO-CORE
func IsIndexSetInBitmap(index uint32, bitmap []byte) bool {
	indexOutOfBounds := index >= uint32(len(bitmap))*8
	if indexOutOfBounds {
		return false
	}

	bytePos := index / 8
	byteInMap := bitmap[bytePos]
	bitPos := index % 8
	mask := uint8(1 << bitPos)
	return (byteInMap & mask) != 0
}
