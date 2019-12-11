package track

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

func (bbt baseBlockTrack) AddHeader(header data.HeaderHandler, hash []byte) {
	bbt.addHeader(header, hash)
}
