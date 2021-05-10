package interceptedBlocks

import (
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
)

// TryCreateHeaderV2 -
func TryCreateHeaderV2(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.HeaderHandler, error) {
	return tryCreateHeaderV2(marshalizer, hdrBuff)
}

// TryCreateHeaderV1 -
func TryCreateHeaderV1(marshalizer marshal.Marshalizer, hdrBuff []byte) (data.HeaderHandler, error) {
	return tryCreateHeaderV1(marshalizer, hdrBuff)
}
