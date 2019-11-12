package metachain

import "github.com/ElrondNetwork/elrond-go/data/block"

type pendingMiniBlock struct {
}

func (p *pendingMiniBlock) PendingMiniBlocks() []block.MiniBlockHeader {
	panic("implement me")
}

func (p *pendingMiniBlock) AddMiniBlockHeader([]block.MiniBlockHeader) {
	panic("implement me")
}

func (p *pendingMiniBlock) IsInterfaceNil() bool {
	return p == nil
}
