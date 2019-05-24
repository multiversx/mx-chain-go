package splitters

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/core"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
)

const minimumMaxPacketSizeInBytes = 1

// SliceSplitter can split a large slice of byte slices in chunks <= maxPacketSize
// If one element still exceeds maxPacketSize, it will be sent alone
// It does the marshaling of the resulted (smaller) slice of byte slices
type SliceSplitter struct {
	marshalizer marshal.Marshalizer
}

// NewSliceSplitter creates a new SliceSplitter instance
func NewSliceSplitter(
	marshalizer marshal.Marshalizer,
) (*SliceSplitter, error) {

	if marshalizer == nil {
		return nil, core.ErrNilMarshalizer
	}

	return &SliceSplitter{
		marshalizer: marshalizer,
	}, nil
}

// SendDataInChunks splits the provided data into smaller chunks, calling for each chunk sendHandler func
func (ss *SliceSplitter) SendDataInChunks(data [][]byte, sendHandler func(buff []byte) error, maxPacketSize int) error {
	if maxPacketSize < minimumMaxPacketSizeInBytes {
		return core.ErrInvalidValue
	}
	if sendHandler == nil {
		return core.ErrNilSendHandler
	}

	elements := make([][]byte, 0)
	lastMarshaledElements := make([]byte, 0)
	for i := 0; i < len(data); i++ {
		element := data[i]
		elements = append(elements, element)

		marshaledElements, err := ss.marshalizer.Marshal(elements)
		if err != nil {
			return err
		}

		isSingleElement := len(elements) == 1
		isMarshaledBuffTooLarge := len(marshaledElements) >= maxPacketSize

		if isMarshaledBuffTooLarge {
			if isSingleElement {
				err = sendHandler(marshaledElements)
				if err != nil {
					return err
				}
			} else {
				err = sendHandler(lastMarshaledElements)
				if err != nil {
					return err
				}
				//we need to decrement i as to retest the current element if it is too large and has to sent alone
				//think of it as "with current element the buffer is too large, send last 'good' buffer,
				//step-back and retry the same element alone"
				i--
			}

			elements = make([][]byte, 0)
			lastMarshaledElements = make([]byte, 0)
			continue
		}

		lastMarshaledElements = marshaledElements
	}

	isRemainingDataStillInBuffer := len(lastMarshaledElements) > 0
	if isRemainingDataStillInBuffer {
		err := sendHandler(lastMarshaledElements)
		if err != nil {
			return err
		}
	}

	return nil
}
