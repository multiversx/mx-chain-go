package process

import "github.com/ElrondNetwork/elrond-go-core/data/block"

// ReceiptsHolder holds the content to be saved in (loaded from) the receipts unit
type ReceiptsHolder struct {
	Miniblocks []*block.MiniBlock
}
