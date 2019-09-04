package partitioning

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// SimpleDataPacker can split a large slice of byte slices in chunks <= maxPacketSize
// If one element still exceeds maxPacketSize, it will be returned alone
// It does the marshaling of the resulted (smaller) slice of byte slices
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

	elements := make([][]byte, 0)
	lenElements := 0
	for _, element := range data {
		isBuffToLarge := lenElements+len(element) >= limit
		elementsNotEmpty := len(elements) > 0
		if isBuffToLarge && elementsNotEmpty {
			marshaledElements, _ := sdp.marshalizer.Marshal(elements)
			returningBuff = append(returningBuff, marshaledElements)
			elements = make([][]byte, 0)
			lenElements = 0
		}

		elements = append(elements, element)
		lenElements += len(element)
	}

	if len(elements) > 0 {
		marshaledElements, err := sdp.marshalizer.Marshal(elements)
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
