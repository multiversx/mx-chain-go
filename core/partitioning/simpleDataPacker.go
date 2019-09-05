package partitioning

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// SimpleDataPacker can split a large slice of byte slices in chunks <= maxPacketSize
// If one element still exceeds maxPacketSize, it will be returned alone
// It does the marshaling of the resulted (smaller) slice of byte slices
// This is a simpler version of a data packer that does not marshall in a repetitive manner currentChunk slice
// as the SizeDataPacker does. This limitation is lighter in terms of CPU cycles and memory used but is not as precise
// as SizeDataPacker.
type SimpleDataPacker struct {
	marshalizer marshal.Marshalizer
}

// NewSimpleDataPacker creates a new SizeDataPacker instance
func NewSimpleDataPacker(marshalizer marshal.Marshalizer) (*SimpleDataPacker, error) {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return nil, core.ErrNilMarshalizer
	}

	return &SimpleDataPacker{
		marshalizer: marshalizer,
	}, nil
}

// PackDataInChunks packs the provided data into smaller chunks
// limit is expressed in bytes
func (sdp *SimpleDataPacker) PackDataInChunks(data [][]byte, limit int) ([][]byte, error) {
	if limit < minimumMaxPacketSizeInBytes {
		return nil, core.ErrInvalidValue
	}
	if data == nil {
		return nil, core.ErrNilInputData
	}

	returningBuff := make([][]byte, 0)

	currentChunk := make([][]byte, 0)
	lenChunk := 0
	for _, element := range data {
		isBuffToLarge := lenChunk+len(element) >= limit
		chunkNotEmpty := len(currentChunk) > 0
		if isBuffToLarge && chunkNotEmpty {
			marshaledChunk, _ := sdp.marshalizer.Marshal(currentChunk)
			returningBuff = append(returningBuff, marshaledChunk)
			currentChunk = make([][]byte, 0)
			lenChunk = 0
		}

		currentChunk = append(currentChunk, element)
		lenChunk += len(element)
	}

	if len(currentChunk) > 0 {
		marshaledElements, err := sdp.marshalizer.Marshal(currentChunk)
		if err != nil {
			return nil, err
		}
		returningBuff = append(returningBuff, marshaledElements)
	}

	return returningBuff, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (sdp *SimpleDataPacker) IsInterfaceNil() bool {
	if sdp == nil {
		return true
	}
	return false
}
