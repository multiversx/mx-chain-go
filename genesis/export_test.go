package genesis

import "math/big"

func (g *Genesis) InitialBalances() []*InitialBalance {
	return g.initialBalances
}

func (g *Genesis) SetInitialBalances(initialBalances []*InitialBalance) {
	g.initialBalances = initialBalances
}

func (g *Genesis) SetEntireSupply(entireSupply *big.Int) {
	g.entireSupply = entireSupply
}

func (g *Genesis) Process() error {
	return g.process()
}
