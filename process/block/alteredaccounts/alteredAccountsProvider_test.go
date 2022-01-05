package alteredaccounts

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/indexer"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/require"
)

func TestNewAlteredAccountsProvider(t *testing.T) {
	t.Parallel()

	t.Run("nil shard coordinator", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.ShardCoordinator = nil

		aap, err := NewAlteredAccountsProvider(args)
		require.Nil(t, aap)
		require.Equal(t, ErrNilShardCoordinator, err)
	})

	t.Run("nil address converter", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.AddressConverter = nil

		aap, err := NewAlteredAccountsProvider(args)
		require.Nil(t, aap)
		require.Equal(t, ErrNilPubKeyConverter, err)
	})

	t.Run("nil accounts adapter", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.AccountsDB = nil

		aap, err := NewAlteredAccountsProvider(args)
		require.Nil(t, aap)
		require.Equal(t, ErrNilAccountsDB, err)
	})

	t.Run("nil marshalizer", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.Marshalizer = nil

		aap, err := NewAlteredAccountsProvider(args)
		require.Nil(t, aap)
		require.Equal(t, ErrNilMarshalizer, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		aap, err := NewAlteredAccountsProvider(args)

		require.NotNil(t, aap)
		require.NoError(t, err)
	})
}

func TestAlteredAccountsProvider_ExtractAlteredAccountsFromPool(t *testing.T) {
	t.Parallel()

	t.Run("no transaction, should return empty map", testExtractAlteredAccountsFromPoolNoTransaction)
	t.Run("should return sender shard accounts", testExtractAlteredAccountsFromPoolSenderShard)
	t.Run("should return receiver shard accounts", testExtractAlteredAccountsFromPoolReceiverShard)
	t.Run("should return all addresses in self shard", testExtractAlteredAccountsFromPoolBothSenderAndReceiverShards)
	t.Run("should check data from trie", testExtractAlteredAccountsFromPoolTrieDataChecks)
	t.Run("should include esdt data", testExtractAlteredAccountsFromPoolShouldIncludeESDT)
	t.Run("should include nft data", testExtractAlteredAccountsFromPoolShouldIncludeNFT)
	t.Run("should work when an address has balance changes, esdt and nft", testExtractAlteredAccountsFromPoolAddressHasBalanceChangeEsdtAndfNft)
}

func testExtractAlteredAccountsFromPoolNoTransaction(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{})
	require.NoError(t, err)
	require.Empty(t, res)
}

func testExtractAlteredAccountsFromPoolSenderShard(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if strings.HasPrefix(string(address), "sender") {
				return 0
			}

			return 1
		},
		SelfIDCalled: func() uint32 {
			return 0
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Txs: map[string]data.TransactionHandler{
			"hash0": &transaction.Transaction{
				SndAddr: []byte("sender shard - tx0"),
				RcvAddr: []byte("receiver shard - tx0"),
			},
			"hash1": &transaction.Transaction{
				SndAddr: []byte("sender shard - tx1"),
				RcvAddr: []byte("receiver shard - tx1"),
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "sender"))
	}
}

func testExtractAlteredAccountsFromPoolReceiverShard(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if strings.HasPrefix(string(address), "sender") {
				return 0
			}

			return 1
		},
		SelfIDCalled: func() uint32 {
			return 1
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Txs: map[string]data.TransactionHandler{
			"hash0": &transaction.Transaction{
				SndAddr: []byte("sender shard - tx0"),
				RcvAddr: []byte("receiver shard - tx0"),
			},
			"hash1": &transaction.Transaction{
				SndAddr: []byte("sender shard - tx1"),
				RcvAddr: []byte("receiver shard - tx1"),
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "receiver"))
	}
}

func testExtractAlteredAccountsFromPoolBothSenderAndReceiverShards(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			shardIdChar := string(address)[5]

			return uint32(shardIdChar - '0')
		},
		SelfIDCalled: func() uint32 {
			return 0
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Txs: map[string]data.TransactionHandler{
			"hash0": &transaction.Transaction{ // intra-shard 0, different addresses
				SndAddr: []byte("shard0 addr - tx0"),
				RcvAddr: []byte("shard0 addr 2 - tx0"),
			},
			"hash1": &transaction.Transaction{ // intra-shard 0, same addresses
				SndAddr: []byte("shard0 addr 3 - tx1"),
				RcvAddr: []byte("shard0 addr 3 - tx1"),
			},
			"hash2": &transaction.Transaction{ // cross-shard, sender in shard 0
				SndAddr: []byte("shard0 addr - tx2"),
				RcvAddr: []byte("shard1 - tx2"),
			},
			"hash3": &transaction.Transaction{ // cross-shard, receiver in shard 0
				SndAddr: []byte("shard1 addr - tx3"),
				RcvAddr: []byte("shard0 addr - tx3"),
			},
			"hash4": &transaction.Transaction{ // cross-shard, no address in shard 0
				SndAddr: []byte("shard2 addr - tx4"),
				RcvAddr: []byte("shard2 addr - tx3"),
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 5, len(res))

	for key := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "shard0"))
	}
}

func testExtractAlteredAccountsFromPoolTrieDataChecks(t *testing.T) {
	t.Parallel()

	receiverInSelfShard := "receiver in shard 1"
	expectedBalance := big.NewInt(37)
	args := getMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if strings.HasPrefix(string(address), "sender") {
				return 0
			}

			return 1
		},
		SelfIDCalled: func() uint32 {
			return 1
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{
				Balance: expectedBalance,
			}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Txs: map[string]data.TransactionHandler{
			"hash0": &transaction.Transaction{
				SndAddr: []byte("sender in shard 0"),
				RcvAddr: []byte(receiverInSelfShard),
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedAddressKey := args.AddressConverter.Encode([]byte(receiverInSelfShard))
	actualAccount, found := res[expectedAddressKey]
	require.True(t, found)
	require.Equal(t, expectedAddressKey, actualAccount.Address)
	require.Equal(t, expectedBalance.String(), actualAccount.Balance)
}

func testExtractAlteredAccountsFromPoolShouldIncludeESDT(t *testing.T) {
	t.Parallel()

	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
	}
	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{
				RetrieveValueFromDataTrieTrackerCalled: func(_ []byte) ([]byte, error) {
					tokenBytes, _ := args.Marshalizer.Marshal(expectedToken)
					return tokenBytes, nil
				},
			}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event {
						{
							Address: []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
							},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	encodedAddr := hex.EncodeToString([]byte("addr"))
	require.Equal(t, expectedToken.Value.String(), res[encodedAddr].Tokens[0].Balance)
}

func testExtractAlteredAccountsFromPoolShouldIncludeNFT(t *testing.T) {
	t.Parallel()

	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 38,
		},
	}
	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{
				RetrieveValueFromDataTrieTrackerCalled: func(_ []byte) ([]byte, error) {
					tokenBytes, _ := args.Marshalizer.Marshal(expectedToken)
					return tokenBytes, nil
				},
			}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event {
						{
							Address: []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(38).Bytes(),
							},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	encodedAddr := hex.EncodeToString([]byte("addr"))
	require.Equal(t, expectedToken.Value.String(), res[encodedAddr].Tokens[0].Balance)
}

func testExtractAlteredAccountsFromPoolAddressHasBalanceChangeEsdtAndfNft(t *testing.T) {
	t.Parallel()

	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 38,
		},
	}
	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{
				RetrieveValueFromDataTrieTrackerCalled: func(_ []byte) ([]byte, error) {
					tokenBytes, _ := args.Marshalizer.Marshal(expectedToken)
					return tokenBytes, nil
				},
			}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&indexer.Pool{
		Txs: map[string]data.TransactionHandler{
			"hash0": &transaction.Transaction{
				SndAddr: []byte("addr"),
			},
		},
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event {
						{
							Address: []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("esdt"),
							},
						},
						{
							Address: []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								[]byte("nft"),
								big.NewInt(38).Bytes(),
							},
						},
					},
				},
			},
		},
	})
	require.NoError(t, err)

	encodedAddr := hex.EncodeToString([]byte("addr"))
	require.Equal(t, 2, len(res[encodedAddr].Tokens))
}

func getMockArgs() ArgsAlteredAccountsProvider {
	return ArgsAlteredAccountsProvider{
		ShardCoordinator: &testscommon.ShardsCoordinatorMock{},
		AddressConverter: &testscommon.PubkeyConverterMock{},
		Marshalizer:      &testscommon.MarshalizerMock{},
		AccountsDB:       &state.AccountsStub{},
	}
}
