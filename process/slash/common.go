package slash

import "github.com/ElrondNetwork/elrond-go-core/data"

// HeaderInfoList defines a list of data.HeaderInfoHandler
type HeaderInfoList []data.HeaderInfoHandler

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
