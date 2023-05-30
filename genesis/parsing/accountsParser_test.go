package parsing_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	coreData "github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	scrData "github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	transactionData "github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/genesis/mock"
	"github.com/multiversx/mx-chain-go/genesis/parsing"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockInitialAccount() *data.InitialAccount {
	return &data.InitialAccount{
		Address:      "0001",
		Supply:       big.NewInt(5),
		Balance:      big.NewInt(1),
		StakingValue: big.NewInt(2),
		Delegation: &data.DelegationData{
			Address: "0002",
			Value:   big.NewInt(2),
		},
	}
}

func createMockHexPubkeyConverter() *testscommon.PubkeyConverterStub {
	return &testscommon.PubkeyConverterStub{
		DecodeCalled: func(humanReadable string) ([]byte, error) {
			return hex.DecodeString(humanReadable)
		},
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			return hex.EncodeToString(pkBytes)
		},
	}
}

func createMockAccountsParserArgs() genesis.AccountsParserArgs {
	return genesis.AccountsParserArgs{
		GenesisFilePath: "./testdata/genesis_ok.json",
		EntireSupply:    big.NewInt(1),
		MinterAddress:   "",
		PubkeyConverter: createMockHexPubkeyConverter(),
		KeyGenerator:    &mock.KeyGeneratorStub{},
		Hasher:          &hashingMocks.HasherMock{},
		Marshalizer:     &mock.MarshalizerMock{},
	}
}

func createSimpleInitialAccount(address string, balance int64) *data.InitialAccount {
	return &data.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(balance),
		Balance:      big.NewInt(balance),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: "",
			Value:   big.NewInt(0),
		},
	}
}

func createDelegatedInitialAccount(address string, delegatedBytes []byte, delegatedBalance int64) *data.InitialAccount {
	ia := &data.InitialAccount{
		Address:      address,
		Supply:       big.NewInt(delegatedBalance),
		Balance:      big.NewInt(0),
		StakingValue: big.NewInt(0),
		Delegation: &data.DelegationData{
			Address: hex.EncodeToString(delegatedBytes),
			Value:   big.NewInt(delegatedBalance),
		},
	}
	ia.SetAddressBytes(delegatedBytes)

	return ia
}

func TestNewAccountsParser_NilEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.EntireSupply = nil

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrNilEntireSupply))
}

func TestNewAccountsParser_ZeroEntireBalanceShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.EntireSupply = big.NewInt(0)

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrInvalidEntireSupply))
}

func TestNewAccountsParser_BadFilenameShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "inexistent file"

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.NotNil(t, err)
}

func TestNewAccountsParser_NilPubkeyConverterShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "inexistent file"
	args.PubkeyConverter = nil

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilPubkeyConverter, err)
}

func TestNewAccountsParser_NilKeyGeneratorShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "inexistent file"
	args.KeyGenerator = nil

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilKeyGenerator, err)
}

func TestNewAccountsParser_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "inexistent file"
	args.Hasher = nil

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilHasher, err)
}

func TestNewAccountsParser_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "inexistent file"
	args.Marshalizer = nil

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.Equal(t, genesis.ErrNilMarshalizer, err)
}

func TestNewAccountsParser_WrongMinterAddressFormatShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.MinterAddress = "wrongaddressformat"

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestNewAccountsParser_BadJsonShouldErr(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.GenesisFilePath = "testdata/genesis_bad.json"

	ap, err := parsing.NewAccountsParser(args)

	assert.True(t, check.IfNil(ap))
	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestNewAccountsParser_ShouldWork(t *testing.T) {
	t.Parallel()

	args := createMockAccountsParserArgs()
	args.EntireSupply = big.NewInt(30)

	ap, err := parsing.NewAccountsParser(args)

	assert.False(t, check.IfNil(ap))
	assert.Nil(t, err)
	assert.Equal(t, 6, len(ap.InitialAccounts()))
}

//------- process

func TestAccountsParser_ProcessEmptyAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = ""
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyAddress))
}

func TestAccountsParser_ProcessInvalidAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = "invalid address"
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidAddress))
}

func TestAccountsParser_ProcessInvalidPublicKeyShouldErr(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("expected error")
	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ap.SetKeyGenerator(&mock.KeyGeneratorStub{
		CheckPublicKeyValidCalled: func(b []byte) error {
			return expectedErr
		},
	})
	ib := createMockInitialAccount()
	ib.Address = "00"
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidPubKey))
}

func TestAccountsParser_ProcessEmptyDelegationAddressButWithBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = ""
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrEmptyDelegationAddress))
}

func TestAccountsParser_ProcessInvalidDelegationAddressShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Address = "invalid address"
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()

	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationAddress))
}

func TestAccountsParser_ProcessInvalidSupplyShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))

	ib.Supply = big.NewInt(0)

	err = ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidSupply))
}

func TestAccountsParser_ProcessInvalidBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Balance = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidBalance))
}

func TestAccountsParser_ProcessInvalidStakingBalanceShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.StakingValue = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidStakingBalance))
}

func TestAccountsParser_ProcessInvalidDelegationValueShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Delegation.Value = big.NewInt(-1)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrInvalidDelegationValue))
}

func TestAccountsParser_ProcessSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Supply = big.NewInt(4)
	ap.SetInitialAccounts([]*data.InitialAccount{ib})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrSupplyMismatch))
}

func TestAccountsParser_ProcessDuplicatesShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib1 := createMockInitialAccount()
	ib2 := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib1, ib2})

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrDuplicateAddress))
}

func TestAccountsParser_ProcessEntireSupplyMismatchShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(4))

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrEntireSupplyMismatch))
}

func TestAccountsParser_AddressIsSmartContractShouldErr(t *testing.T) {
	t.Parallel()

	addr := strings.Repeat("0", (core.NumInitCharactersForScAddress+1)*2)
	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ib.Address = addr
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(4))

	err := ap.Process()
	assert.True(t, errors.Is(err, genesis.ErrAddressIsSmartContract))
}

func TestAccountsParser_ProcessShouldWork(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib := createMockInitialAccount()
	ap.SetInitialAccounts([]*data.InitialAccount{ib})
	ap.SetEntireSupply(big.NewInt(5))

	err := ap.Process()
	assert.Nil(t, err)
}

//------- InitialAccountsSplitOnAddressesShards

func TestAccountsParser_InitialAccountsSplitOnAddressesShardsNilShardCoordinatorShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ibs, err := ap.InitialAccountsSplitOnAddressesShards(
		nil,
	)

	assert.Nil(t, ibs)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestAccountsParser_InitialAccountsSplitOnAddressesShards(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
		createSimpleInitialAccount("0000", balance),
		createSimpleInitialAccount("0101", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)
	err := ap.Process()
	require.Nil(t, err)

	threeSharder := &mock.ShardCoordinatorMock{
		NumOfShards: 3,
		SelfShardId: 0,
	}
	ibsSplit, err := ap.InitialAccountsSplitOnAddressesShards(
		threeSharder,
	)

	assert.Nil(t, err)
	require.Equal(t, 3, len(ibsSplit))
	assert.Equal(t, 2, len(ibsSplit[1]))
}

func TestAccountsParser_GetInitialAccountsForDelegated(t *testing.T) {
	t.Parallel()

	addr1 := "1000"
	addr2 := "2000"
	delegatedUpon := int64(78)

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	ib1 := createDelegatedInitialAccount("0001", []byte(addr1), delegatedUpon)
	ib2 := createDelegatedInitialAccount("0002", []byte(addr1), delegatedUpon)
	ib3 := createDelegatedInitialAccount("0003", []byte(addr2), delegatedUpon)

	ap.SetEntireSupply(big.NewInt(3 * delegatedUpon))
	ap.SetInitialAccounts([]*data.InitialAccount{ib1, ib2, ib3})

	err := ap.Process()
	require.Nil(t, err)

	list := ap.GetInitialAccountsForDelegated([]byte(addr1))
	require.Equal(t, 2, len(list))
	//order is important
	assert.Equal(t, ib1, list[0])
	assert.Equal(t, ib2, list[1])
	delegated := ap.GetTotalStakedForDelegationAddress(hex.EncodeToString([]byte(addr1)))
	assert.Equal(t, big.NewInt(delegatedUpon*2), delegated)

	list = ap.GetInitialAccountsForDelegated([]byte(addr2))
	require.Equal(t, 1, len(list))
	assert.Equal(t, ib3, list[0])
	delegated = ap.GetTotalStakedForDelegationAddress(hex.EncodeToString([]byte(addr2)))
	assert.Equal(t, big.NewInt(delegatedUpon), delegated)

	list = ap.GetInitialAccountsForDelegated([]byte("not delegated"))
	require.Equal(t, 0, len(list))
	delegated = ap.GetTotalStakedForDelegationAddress(hex.EncodeToString([]byte("not delegated")))
	assert.Equal(t, big.NewInt(0), delegated)
}

//------- GetMintTransactions

func TestAccountsParser_GenerateInitialTransactionsShouldErr(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	miniBlocks, txsPoolPerShard, err := ap.GenerateInitialTransactions(
		nil,
		nil,
	)

	assert.Nil(t, miniBlocks)
	assert.Nil(t, txsPoolPerShard)
	assert.Equal(t, genesis.ErrNilShardCoordinator, err)
}

func TestAccountsParser_getShardIDs(t *testing.T) {
	t.Parallel()

	sharder := &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	shardIDs := parsing.GetShardIDs(sharder)
	assert.Equal(t, 3, len(shardIDs))
}

func TestAccountsParser_createMintTransaction(t *testing.T) {
	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)

	err := ap.Process()
	require.Nil(t, err)

	ia := ap.InitialAccounts()

	tx := ap.CreateMintTransaction(ia[0], uint64(0))
	assert.Equal(t, uint64(0), tx.GetNonce())
	assert.Equal(t, ia[0].AddressBytes(), tx.GetRcvAddr())
	assert.Equal(t, ia[0].GetSupply(), tx.GetValue())
	assert.Equal(t, []byte("erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3"), tx.GetSndAddr())
	assert.Equal(t, []byte(common.GenesisTxSignatureString), tx.GetSignature())
	assert.Equal(t, uint64(0), tx.GetGasLimit())
	assert.Equal(t, uint64(0), tx.GetGasPrice())
}

func TestAccountsParser_createMintTransactions(t *testing.T) {
	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
		createSimpleInitialAccount("0000", balance),
		createSimpleInitialAccount("0103", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)

	err := ap.Process()
	require.Nil(t, err)

	txs := ap.CreateMintTransactions()
	assert.Equal(t, 4, len(txs))
	assert.Equal(t, uint64(0), txs[0].GetNonce())
	assert.Equal(t, uint64(3), txs[3].GetNonce())
}

func TestAccountsParser_createMiniBlocks(t *testing.T) {
	t.Parallel()

	shardIDs := []uint32{0, 1}

	miniBlocks := parsing.CreateMiniBlocks(shardIDs, block.TxBlock)
	assert.Equal(t, 4, len(miniBlocks))
}

func TestAccountsParser_setScrsTxsPool(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())

	sharder := &mock.ShardCoordinatorMock{
		NumOfShards: 1,
		SelfShardId: 0,
	}

	scrsTxs := make(map[string]coreData.TransactionHandler)
	scrsTxs["hash"] = &scrData.SmartContractResult{
		Nonce:    1,
		RcvAddr:  bytes.Repeat([]byte{0}, 32),
		SndAddr:  bytes.Repeat([]byte{1}, 32),
		GasLimit: 10000,
	}

	indexingDataMap := make(map[uint32]*genesis.IndexingData)
	indexingData := &genesis.IndexingData{
		ScrsTxs: scrsTxs,
	}
	for i := uint32(0); i < sharder.NumOfShards; i++ {
		indexingDataMap[i] = indexingData
	}

	txsPoolPerShard := make(map[uint32]*outport.TransactionPool)
	for i := uint32(0); i < sharder.NumOfShards; i++ {
		txsPoolPerShard[i] = &outport.TransactionPool{
			SmartContractResults: map[string]*outport.SCRInfo{},
		}
	}

	ap.SetScrsTxsPool(sharder, indexingDataMap, txsPoolPerShard)
	assert.Equal(t, 1, len(txsPoolPerShard))
	assert.Equal(t, uint64(0), txsPoolPerShard[0].SmartContractResults["hash"].SmartContractResult.GetGasLimit())
	assert.Equal(t, uint64(1), txsPoolPerShard[0].SmartContractResults["hash"].SmartContractResult.GetNonce())
}

func TestAccountsParser_GenerateInitialTransactionsTxsPool(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)

	err := ap.Process()
	require.Nil(t, err)

	sharder := &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	indexingDataMap := make(map[uint32]*genesis.IndexingData)
	indexingData := &genesis.IndexingData{
		DelegationTxs:      make([]coreData.TransactionHandler, 0),
		ScrsTxs:            make(map[string]coreData.TransactionHandler),
		StakingTxs:         make([]coreData.TransactionHandler, 0),
		DeploySystemScTxs:  make([]coreData.TransactionHandler, 0),
		DeployInitialScTxs: make([]coreData.TransactionHandler, 0),
	}
	for i := uint32(0); i < sharder.NumOfShards; i++ {
		indexingDataMap[i] = indexingData
	}

	miniBlocks, txsPoolPerShard, err := ap.GenerateInitialTransactions(sharder, indexingDataMap)
	require.Nil(t, err)

	assert.Equal(t, 2, len(miniBlocks))

	assert.Equal(t, 3, len(txsPoolPerShard))
	assert.Equal(t, 1, len(txsPoolPerShard[0].Transactions))
	assert.Equal(t, 1, len(txsPoolPerShard[1].Transactions))
	assert.Equal(t, len(ibs), len(txsPoolPerShard[core.MetachainShardId].Transactions))
	assert.Equal(t, 0, len(txsPoolPerShard[0].SmartContractResults))
	assert.Equal(t, 0, len(txsPoolPerShard[1].SmartContractResults))
	assert.Equal(t, 0, len(txsPoolPerShard[core.MetachainShardId].SmartContractResults))

	for _, tx := range txsPoolPerShard[1].Transactions {
		assert.Equal(t, ibs[0].GetSupply(), tx.Transaction.GetValue())
		assert.Equal(t, ibs[0].AddressBytes(), tx.Transaction.GetRcvAddr())
	}

	for _, tx := range txsPoolPerShard[0].Transactions {
		assert.Equal(t, ibs[1].GetSupply(), tx.Transaction.GetValue())
		assert.Equal(t, ibs[1].AddressBytes(), tx.Transaction.GetRcvAddr())
	}

}

func TestAccountsParser_GenerateInitialTransactionsZeroGasLimitShouldWork(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := []*data.InitialAccount{
		createSimpleInitialAccount("0001", balance),
		createSimpleInitialAccount("0002", balance),
		createSimpleInitialAccount("0000", balance),
		createSimpleInitialAccount("0103", balance),
	}

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)

	err := ap.Process()
	require.Nil(t, err)

	sharder := &mock.ShardCoordinatorMock{
		NumOfShards: 2,
		SelfShardId: 0,
	}

	indexingDataMap := make(map[uint32]*genesis.IndexingData)
	_, txsPoolPerShard, err := ap.GenerateInitialTransactions(sharder, indexingDataMap)
	require.Nil(t, err)

	for i := uint32(0); i < sharder.NumberOfShards(); i++ {
		for _, tx := range txsPoolPerShard[i].Transactions {
			assert.Equal(t, uint64(0), tx.Transaction.GetGasLimit())
		}
	}
}

func TestAccountsParser_GenerateInitialTransactionsVerifyTxsHashes(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	balance := int64(1)
	ibs := make([]*data.InitialAccount, 0)

	ap.SetEntireSupply(big.NewInt(int64(len(ibs)) * balance))
	ap.SetInitialAccounts(ibs)

	err := ap.Process()
	require.Nil(t, err)

	sharder := &mock.ShardCoordinatorMock{
		NumOfShards: 1,
		SelfShardId: 0,
	}

	tx := &transactionData.Transaction{
		Nonce:     0,
		GasPrice:  0,
		GasLimit:  0,
		Signature: []byte(common.GenesisTxSignatureString),
	}
	hashHex := "cef3536e36ae01d3c84c97b0e6fae577f34c12c0cfdb51a04a2668afd5f5efe7"
	txHash, err := hex.DecodeString(hashHex)
	require.Nil(t, err)

	indexingDataMap := make(map[uint32]*genesis.IndexingData)
	indexingData := &genesis.IndexingData{
		DelegationTxs: []coreData.TransactionHandler{tx},
	}
	indexingDataMap[0] = indexingData

	miniBlocks, txsPoolPerShard, err := ap.GenerateInitialTransactions(sharder, indexingDataMap)
	require.Nil(t, err)

	assert.Equal(t, 1, len(miniBlocks))
	assert.Equal(t, 2, len(txsPoolPerShard))
	assert.Equal(t, 1, len(txsPoolPerShard[0].Transactions))

	for hashString, v := range txsPoolPerShard[0].Transactions {
		assert.Equal(t, txHash, []byte(hashString))
		assert.Equal(t, tx, v.Transaction)
	}
}

func TestAccountsParser_GenesisMintingAddress(t *testing.T) {
	t.Parallel()

	ap := parsing.NewTestAccountsParser(createMockHexPubkeyConverter())
	addr := ap.GenesisMintingAddress()
	assert.Equal(t, hex.EncodeToString([]byte("erd17rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rc0pu8s7rcqqkhty3")), addr)
}
