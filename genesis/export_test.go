package genesis

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
)

func (g *Genesis) InitialAccounts() []*InitialAccount {
	return g.initialAccounts
}

func (g *Genesis) SetInitialAccounts(initialAccounts []*InitialAccount) {
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
		initialAccounts: make([]*InitialAccount, 0),
	}
}
