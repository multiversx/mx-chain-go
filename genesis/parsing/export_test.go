package parsing

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go/core"
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

func (ap *accountsParser) SetPukeyConverter(pubkeyConverter core.PubkeyConverter) {
	ap.pubkeyConverter = pubkeyConverter
}

func NewTestAccountsParser(pubkeyConverter core.PubkeyConverter) *accountsParser {
	return &accountsParser{
		pubkeyConverter: pubkeyConverter,
		initialAccounts: make([]*data.InitialAccount, 0),
	}
}

func NewTestSmartContractsParser(pubkeyConverter core.PubkeyConverter) *smartContractParser {
	scp := &smartContractParser{
		pubkeyConverter:       pubkeyConverter,
		initialSmartContracts: make([]*data.InitialSmartContract, 0),
	}
	//mock implementation, assumes the files are present
	scp.checkForFileHandler = func(filename string) error {
		return nil
	}

	return scp
}

func (scp *smartContractParser) SetInitialSmartContracts(initialSmartContracts []*data.InitialSmartContract) {
	scp.initialSmartContracts = initialSmartContracts
}

func (scp *smartContractParser) Process() error {
	return scp.process()
}

func (scp *smartContractParser) SetFileHandler(handler func(string) error) {
	scp.checkForFileHandler = handler
}
