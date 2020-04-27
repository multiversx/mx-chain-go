package parsing

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
)

func (ap *accountsParser) SetInitialAccounts(initialAccounts []*data.InitialAccount) {
	ap.initialAccounts = initialAccounts
}

func (ap *accountsParser) SetEntireSupply(entireSupply *big.Int) {
	ap.entireSupply = entireSupply
}

func (ap *accountsParser) Process() error {
	return ap.process()
}

func (ap *accountsParser) SetPukeyConverter(pubkeyConverter state.PubkeyConverter) {
	ap.pubkeyConverter = pubkeyConverter
}

func NewTestAccountsParser(pubkeyConverter state.PubkeyConverter) *accountsParser {
	return &accountsParser{
		pubkeyConverter: pubkeyConverter,
		initialAccounts: make([]*data.InitialAccount, 0),
	}
}
