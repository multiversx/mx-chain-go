package parsing_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/stretchr/testify/require"
)

func TestNewSovereignAccountsParser(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		accParser := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
		sovAccParser, err := parsing.NewSovereignAccountsParser(accParser)
		require.Nil(t, err)
		require.False(t, sovAccParser.IsInterfaceNil())
	})

	t.Run("nil accounts parser, should return error", func(t *testing.T) {
		sovAccParser, err := parsing.NewSovereignAccountsParser(nil)
		require.Equal(t, parsing.ErrNilAccountsParser, err)
		require.Nil(t, sovAccParser)
	})
}

func txFoundForInitialAccount(tx *outport.TxInfo, accounts []*data.InitialAccount) bool {
	for _, acc := range accounts {
		if bytes.Equal(tx.GetTransaction().GetRcvAddr(), acc.AddressBytes()) &&
			tx.GetTransaction().GetValue().Cmp(acc.GetSupply()) == 0 {
			return true
		}
	}

	return false
}

func requireInitialTxMapContainsAccounts(t *testing.T, txMap map[string]*outport.TxInfo, accounts []*data.InitialAccount) {
	for _, tx := range txMap {
		if !txFoundForInitialAccount(tx, accounts) {
			require.Fail(t, "initial tx not found for accounts")
		}
	}
}

func TestSovereignAccountsParser_GenerateInitialTransactions(t *testing.T) {
	t.Parallel()

	accParser := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	sovAccParser, _ := parsing.NewSovereignAccountsParser(accParser)
	balance := int64(1)
	accounts := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
	}

	sovAccParser.SetEntireSupply(big.NewInt(int64(len(accounts)) * balance))
	sovAccParser.SetInitialAccounts(accounts)

	err := sovAccParser.Process()
	require.Nil(t, err)

	shardCoordinator := sharding.NewSovereignShardCoordinator(core.SovereignChainShardId)
	indexingDataMap := map[uint32]*genesis.IndexingData{
		core.SovereignChainShardId: {
			DelegationTxs:      make([]coreData.TransactionHandler, 0),
			ScrsTxs:            make(map[string]coreData.TransactionHandler),
			StakingTxs:         make([]coreData.TransactionHandler, 0),
			DeploySystemScTxs:  make([]coreData.TransactionHandler, 0),
			DeployInitialScTxs: make([]coreData.TransactionHandler, 0),
		},
	}
	miniBlocks, txsPoolPerShard, err := sovAccParser.GenerateInitialTransactions(nil, indexingDataMap)
	require.Equal(t, core.ErrNilShardCoordinator, err)
	require.Nil(t, miniBlocks)
	require.Nil(t, txsPoolPerShard)

	miniBlocks, txsPoolPerShard, err = sovAccParser.GenerateInitialTransactions(shardCoordinator, indexingDataMap)
	require.Nil(t, err)

	require.Equal(t, 1, len(miniBlocks))
	require.Equal(t, 1, len(txsPoolPerShard))
	require.Equal(t, 2, len(txsPoolPerShard[core.SovereignChainShardId].Transactions))
	require.Equal(t, 0, len(txsPoolPerShard[core.SovereignChainShardId].SmartContractResults))
	requireInitialTxMapContainsAccounts(t, txsPoolPerShard[core.SovereignChainShardId].Transactions, accounts)
}
