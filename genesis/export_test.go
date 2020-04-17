package genesis

import "math/big"

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
