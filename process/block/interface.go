package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type blockProcessor interface {
	CreateNewHeader(round uint64) data.HeaderHandler
}
