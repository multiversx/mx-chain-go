package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type blockProcessor interface {
	decodeBlockHeader(dta []byte) data.HeaderHandler
}
