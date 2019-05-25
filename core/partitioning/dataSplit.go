package partitioning

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
)

const minimumMaxPacketNum = 1

// DataSplit can split a large slice of byte slices in chunks with len(chunks) <= limit
// It does not marshal the data
type DataSplit struct {
}

// SplitDataInChunks splits the provided data into smaller chunks
// limit is expressed in number of elements
func (ds *DataSplit) SplitDataInChunks(data [][]byte, limit int) ([][][]byte, error) {
	if limit < minimumMaxPacketNum {
		return nil, core.ErrInvalidValue
	}
	if data == nil {
		return nil, core.ErrNilInputData
	}

	returningBuff := make([][][]byte, 0)

	elements := make([][]byte, 0)
	for idx, element := range data {
		elements = append(elements, element)

		if (idx+1)%limit == 0 {
			returningBuff = append(returningBuff, elements)
			elements = make([][]byte, 0)
		}
	}

	if len(elements) > 0 {
		returningBuff = append(returningBuff, elements)
	}

	return returningBuff, nil
}
