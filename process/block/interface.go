package block

import (
	"github.com/ElrondNetwork/elrond-go-core/data"
)

type blockProcessor interface {
	removeStartOfEpochBlockDataFromPools(headerHandler data.HeaderHandler, bodyHandler data.BodyHandler) error
}
