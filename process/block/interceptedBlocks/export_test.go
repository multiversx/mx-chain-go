package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// CreateHeaderV2 -
func CreateHeaderV2(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.HeaderHandler, error) {
	return createHeaderV2(marshalizer, hdrBuff)
}

// CreateHeaderV1 -
func CreateHeaderV1(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.HeaderHandler, error) {
	return createHeaderV1(marshalizer, hdrBuff)
}
