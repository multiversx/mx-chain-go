package genesis

import (
	"math/big"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// IndexingData specifies transactions sets that will be used for indexing
type IndexingData struct {
	DelegationTxs      []data.TransactionHandler
	ScrsTxs            map[string]data.TransactionHandler
	StakingTxs         []data.TransactionHandler
	DeploySystemScTxs  []data.TransactionHandler
	DeployInitialScTxs []data.TransactionHandler
}

// AccountsParserArgs holds all dependencies required by the accounts parser in order to create new instances
type AccountsParserArgs struct {
	InitialAccounts []InitialAccountHandler
	EntireSupply    *big.Int
	MinterAddress   string
	PubkeyConverter core.PubkeyConverter
	KeyGenerator    crypto.KeyGenerator
	Hasher          hashing.Hasher
	Marshalizer     marshal.Marshalizer
}
