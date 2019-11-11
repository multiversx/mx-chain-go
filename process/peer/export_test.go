package peer

import (
	"github.com/ElrondNetwork/elrond-go/sharding"
	"math/big"
)

func (p *validatorStatistics) SaveInitialState(in []*sharding.InitialNode, stakeValue *big.Int) error {
	return p.saveInitialState(in, stakeValue)
}
