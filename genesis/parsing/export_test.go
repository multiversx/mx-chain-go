package parsing

import (
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	transactionData "github.com/ElrondNetwork/elrond-go-core/data/transaction"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/genesis/data"
	"github.com/ElrondNetwork/elrond-go/genesis/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
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

func (ap *accountsParser) GenerateGenesisMintTransactions(shardCoordinator sharding.Coordinator) map[uint32][]*transactionData.Transaction {
	return ap.generateMintTransactionsPerShards(shardCoordinator)
}

func (ap *accountsParser) GenerateGenesisMiniBlocks(txsPerShards map[uint32][]*transactionData.Transaction) ([]*block.MiniBlock, error) {
	return ap.generateInShardMiniBlocks(txsPerShards)
}

func (ap *accountsParser) CalculateTxHashes(txs []*transactionData.Transaction) ([][]byte, error) {
	return ap.calculateTxHashes(txs)
}

func (ap *accountsParser) SetPukeyConverter(pubkeyConverter core.PubkeyConverter) {
	ap.pubkeyConverter = pubkeyConverter
}

func (ap *accountsParser) SetKeyGenerator(keyGen crypto.KeyGenerator) {
	ap.keyGenerator = keyGen
}

func NewTestAccountsParser(pubkeyConverter core.PubkeyConverter) *accountsParser {
	return &accountsParser{
		pubkeyConverter: pubkeyConverter,
		initialAccounts: make([]*data.InitialAccount, 0),
		keyGenerator:    &mock.KeyGeneratorStub{},
		hasher:          &mock.HasherMock{},
		marshalizer:     &mock.MarshalizerMock{},
	}
}

func NewTestSmartContractsParser(pubkeyConverter core.PubkeyConverter) *smartContractParser {
	scp := &smartContractParser{
		pubkeyConverter:       pubkeyConverter,
		keyGenerator:          &mock.KeyGeneratorStub{},
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

func (scp *smartContractParser) SetKeyGenerator(keyGen crypto.KeyGenerator) {
	scp.keyGenerator = keyGen
}
