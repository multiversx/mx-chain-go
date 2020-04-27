package parsing

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
)

func (g *Genesis) SetInitialAccounts(initialAccounts []*data.InitialAccount) {
	g.initialAccounts = initialAccounts
}

func (g *Genesis) SetEntireSupply(entireSupply *big.Int) {
	g.entireSupply = entireSupply
}

func (g *Genesis) Process() error {
	return g.process()
}

func (g *Genesis) SetPukeyConverter(pubkeyConverter state.PubkeyConverter) {
	g.pubkeyConverter = pubkeyConverter
}

func NewTestGenesis(pubkeyConverter state.PubkeyConverter) *Genesis {
	return &Genesis{
		pubkeyConverter: pubkeyConverter,
		initialAccounts: make([]*data.InitialAccount, 0),
	}
}
