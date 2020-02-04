package block

import (
	"github.com/ElrondNetwork/elrond-go/data"
)

type blockProcessor interface {
	CreateNewHeader() data.HeaderHandler
}
