package alteredaccounts

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	outportcore "github.com/ElrondNetwork/elrond-go-core/data/outport"
	"github.com/ElrondNetwork/elrond-go-core/data/rewardTx"
	"github.com/ElrondNetwork/elrond-go-core/data/smartContractResult"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/outport/process/alteredaccounts/shared"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/state"
	"github.com/ElrondNetwork/elrond-go/testscommon/trie"
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
		require.Equal(t, errNilShardCoordinator, err)
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

	t.Run("nil esdt data storage handler", func(t *testing.T) {
		t.Parallel()

		args := getMockArgs()
		args.EsdtDataStorageHandler = nil

		aap, err := NewAlteredAccountsProvider(args)
		require.Nil(t, aap)
		require.Equal(t, ErrNilESDTDataStorageHandler, err)
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
	t.Run("should return addresses from scrs, invalid and rewards", testExtractAlteredAccountsFromPoolScrsInvalidRewards)
	t.Run("should check data from trie", testExtractAlteredAccountsFromPoolTrieDataChecks)
	t.Run("should error when casting to vm common user account handler", testExtractAlteredAccountsFromPoolShouldReturnErrorWhenCastingToVmCommonUserAccountHandler)
	t.Run("should include esdt data", testExtractAlteredAccountsFromPoolShouldIncludeESDT)
	t.Run("should include nft data", testExtractAlteredAccountsFromPoolShouldIncludeNFT)
	t.Run("should not include receiver if log if nft create", testExtractAlteredAccountsFromPoolShouldNotIncludeReceiverAddressIfNftCreateLog)
	t.Run("should include receiver from tokens logs", testExtractAlteredAccountsFromPoolShouldIncludeDestinationFromTokensLogsTopics)
	t.Run("should work when an address has balance changes, esdt and nft", testExtractAlteredAccountsFromPoolAddressHasBalanceChangeEsdtAndfNft)
	t.Run("should work when an address has multiple nfts with different nonces", testExtractAlteredAccountsFromPoolAddressHasMultipleNfts)
}

func testExtractAlteredAccountsFromPoolNoTransaction(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{}, shared.AlteredAccountsOptions{})
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
	args.AddressConverter = testscommon.NewPubkeyConverterMock(20)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender shard - tx0  "),
				RcvAddr: []byte("receiver shard - tx0"),
			}, 0, big.NewInt(0)),
			"hash1": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender shard - tx1  "),
				RcvAddr: []byte("receiver shard - tx1"),
			}, 0, big.NewInt(0)),
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key, info := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "sender"))
		require.True(t, info.IsSender)
		require.True(t, info.BalanceChanged)
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
	args.AddressConverter = testscommon.NewPubkeyConverterMock(20)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender shard - tx0  "),
				RcvAddr: []byte("receiver shard - tx0"),
			}, 0, big.NewInt(0)),
			"hash1": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender shard - tx1  "),
				RcvAddr: []byte("receiver shard - tx1"),
			}, 0, big.NewInt(0)),
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key, info := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "receiver"))
		require.True(t, info.BalanceChanged)
		require.False(t, info.IsSender)
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
	args.AddressConverter = testscommon.NewPubkeyConverterMock(19)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{ // intra-shard 0, different addresses
				SndAddr: []byte("shard0 addr - tx0  "),
				RcvAddr: []byte("shard0 addr 2 - tx0"),
			}, 0, big.NewInt(0)),
			"hash1": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{ // intra-shard 0, same addresses
				SndAddr: []byte("shard0 addr 3 - tx1"),
				RcvAddr: []byte("shard0 addr 3 - tx1"),
			}, 0, big.NewInt(0)),
			"hash2": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{ // cross-shard, sender in shard 0
				SndAddr: []byte("shard0 addr - tx2  "),
				RcvAddr: []byte("shard1 - tx2       "),
			}, 0, big.NewInt(0)),
			"hash3": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{ // cross-shard, receiver in shard 0
				SndAddr: []byte("shard1 addr - tx3  "),
				RcvAddr: []byte("shard0 addr - tx3  "),
			}, 0, big.NewInt(0)),
			"hash4": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{ // cross-shard, no address in shard 0
				SndAddr: []byte("shard2 addr - tx4  "),
				RcvAddr: []byte("shard2 addr - tx3  "),
			}, 0, big.NewInt(0)),
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 5, len(res))

	for key := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "shard0"))
	}
	require.Contains(t, res, args.AddressConverter.Encode([]byte("shard0 addr - tx0  ")))
	require.Contains(t, res, args.AddressConverter.Encode([]byte("shard0 addr 2 - tx0")))
	require.Contains(t, res, args.AddressConverter.Encode([]byte("shard0 addr 3 - tx1")))
	require.Contains(t, res, args.AddressConverter.Encode([]byte("shard0 addr - tx2  ")))
	require.Contains(t, res, args.AddressConverter.Encode([]byte("shard0 addr - tx3  ")))
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
	args.AddressConverter = testscommon.NewPubkeyConverterMock(19)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender in shard 0  "),
				RcvAddr: []byte(receiverInSelfShard),
			}, 0, big.NewInt(0)),
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedAddressKey := args.AddressConverter.Encode([]byte(receiverInSelfShard))
	actualAccount, found := res[expectedAddressKey]
	require.True(t, found)
	require.Equal(t, expectedAddressKey, actualAccount.Address)
	require.Equal(t, expectedBalance.String(), actualAccount.Balance)
}

func testExtractAlteredAccountsFromPoolScrsInvalidRewards(t *testing.T) {
	t.Parallel()

	expectedBalance := big.NewInt(37)
	args := getMockArgs()
	args.ShardCoordinator = &testscommon.ShardsCoordinatorMock{
		ComputeIdCalled: func(address []byte) uint32 {
			if strings.Contains(string(address), "shard 0") {
				return 0
			}

			return 1
		},
		SelfIDCalled: func() uint32 {
			return 0
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(container []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{
				Balance: expectedBalance,
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(26)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender in shard 0 - tx 0  "),
			}, 0, big.NewInt(0)),
		},
		Rewards: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash1": outportcore.NewTransactionHandlerWithGasAndFee(&rewardTx.RewardTx{
				RcvAddr: []byte("receiver in shard 0 - tx 1"),
			}, 0, big.NewInt(0)),
		},
		Scrs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash2": outportcore.NewTransactionHandlerWithGasAndFee(&smartContractResult.SmartContractResult{
				SndAddr: []byte("sender in shard 0 - tx 2  "),
				RcvAddr: []byte("receiver in shard 0 - tx 2"),
			}, 0, big.NewInt(0)),
		},
		Invalid: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash3": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender in shard 0 - tx 3  "),
				RcvAddr: []byte("receiver in shard 0 - tx 3"), // receiver for invalid txs should not be included
			}, 0, big.NewInt(0)),
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Len(t, res, 5)
}

func testExtractAlteredAccountsFromPoolShouldReturnErrorWhenCastingToVmCommonUserAccountHandler(t *testing.T) {
	t.Parallel()

	expectedToken := esdt.ESDigitalToken{
		Value:      big.NewInt(37),
		Properties: []byte("ok"),
	}
	args := getMockArgs()
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.UserAccountStub{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
							},
						},
						{
							Address:    []byte("addr"), // other event for the same token, to ensure it isn't added twice
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.True(t, errors.Is(err, errCannotCastToVmCommonUserAccountHandler))
	require.Nil(t, res)
}

func testExtractAlteredAccountsFromPoolShouldIncludeESDT(t *testing.T) {
	t.Parallel()

	expectedToken := esdt.ESDigitalToken{
		Value:      big.NewInt(37),
		Properties: []byte("ok"),
	}
	args := getMockArgs()
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
							},
						},
						{
							Address:    []byte("addr"), // other event for the same token, to ensure it isn't added twice
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr := args.AddressConverter.Encode([]byte("addr"))

	require.Len(t, res, 1)
	require.Len(t, res[encodedAddr].Tokens, 1)
	require.Equal(t, &outportcore.AccountTokenData{
		Identifier: "token0",
		Balance:    expectedToken.Value.String(),
		Nonce:      0,
		Properties: "ok",
		MetaData:   nil,
	}, res[encodedAddr].Tokens[0])
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
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
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
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr := args.AddressConverter.Encode([]byte("addr"))
	require.Equal(t, &outportcore.AccountTokenData{
		Identifier: "token0",
		Balance:    expectedToken.Value.String(),
		Nonce:      expectedToken.TokenMetaData.Nonce,
		MetaData:   expectedToken.TokenMetaData,
	}, res[encodedAddr].Tokens[0])
}

func testExtractAlteredAccountsFromPoolShouldNotIncludeReceiverAddressIfNftCreateLog(t *testing.T) {
	t.Parallel()

	receiverOnDestination := []byte("receiver on destination shard")
	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 38,
		},
	}
	args := getMockArgs()
	args.AddressConverter = testscommon.NewPubkeyConverterMock(26)
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hh": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("sender in shard 0 - tx 1  "),
				RcvAddr: []byte("sender in shard 0 - tx 1  "),
			}, 0, big.NewInt(0)),
		},
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("sender in shard 0 - tx 1  "),
					Events: []*transaction.Event{
						{
							Address:    []byte("sender in shard 0 - tx 1  "),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(38).Bytes(),
								nil,
								receiverOnDestination,
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	require.Len(t, res, 1)
	require.True(t, res["73656e64657220696e2073686172642030202d20747820312020"].Tokens[0].IsNFTCreate)
	require.True(t, res["73656e64657220696e2073686172642030202d20747820312020"].BalanceChanged)
	require.True(t, res["73656e64657220696e2073686172642030202d20747820312020"].IsSender)

	mapKeyToSearch := args.AddressConverter.Encode(receiverOnDestination)
	require.Nil(t, res[mapKeyToSearch])
}

func testExtractAlteredAccountsFromPoolShouldIncludeDestinationFromTokensLogsTopics(t *testing.T) {
	t.Parallel()

	receiverOnDestination := []byte("receiver on destination shard")
	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 38,
		},
	}
	args := getMockArgs()
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(38).Bytes(),
								nil,
								receiverOnDestination,
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	require.Len(t, res, 2)

	mapKeyToSearch := args.AddressConverter.Encode(receiverOnDestination)
	require.Len(t, res[mapKeyToSearch].Tokens, 1)
	require.Equal(t, res[mapKeyToSearch].Tokens[0], &outportcore.AccountTokenData{
		Identifier: "token0",
		Balance:    "37",
		Nonce:      38,
		MetaData: &esdt.MetaData{
			Nonce: 38,
		},
	})
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
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			return &expectedToken, false, nil
		},
	}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("addr"),
			}, 0, big.NewInt(0)),
		},
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("esdt"),
							},
						},
						{
							Address:    []byte("addr"),
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
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr := args.AddressConverter.Encode([]byte("addr"))
	require.Len(t, res[encodedAddr].Tokens, 2)
}

func testExtractAlteredAccountsFromPoolAddressHasMultipleNfts(t *testing.T) {
	t.Parallel()

	expectedToken0 := esdt.ESDigitalToken{
		Value: big.NewInt(37),
	}
	expectedToken1 := esdt.ESDigitalToken{
		Value: big.NewInt(38),
		TokenMetaData: &esdt.MetaData{
			Nonce: 5,
			Name:  []byte("nft-0"),
		},
	}
	expectedToken2 := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 6,
			Name:  []byte("nft-0"),
		},
	}
	args := getMockArgs()
	args.EsdtDataStorageHandler = &testscommon.EsdtStorageHandlerStub{
		GetESDTNFTTokenOnDestinationCalled: func(acnt vmcommon.UserAccountHandler, esdtTokenKey []byte, nonce uint64) (*esdt.ESDigitalToken, bool, error) {
			if strings.Contains(string(esdtTokenKey), "esdttoken") {
				return &expectedToken0, false, nil
			}

			if strings.Contains(string(esdtTokenKey), "nft-0") && nonce == 5 {
				return &expectedToken1, false, nil
			}

			if strings.Contains(string(esdtTokenKey), "nft-0") && nonce == 6 {
				return &expectedToken2, false, nil
			}

			return nil, false, nil
		},
	}
	marshaller := testscommon.MarshalizerMock{}
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			trieMock := trie.DataTrieTrackerStub{
				RetrieveValueCalled: func(key []byte) ([]byte, error) {
					if strings.Contains(string(key), "esdttoken") {
						tokenBytes, _ := marshaller.Marshal(expectedToken0)
						return tokenBytes, nil
					}

					firstNftKey := fmt.Sprintf("%s%s", "nft-0", string(big.NewInt(5).Bytes()))
					if strings.Contains(string(key), firstNftKey) {
						tokenBytes, _ := marshaller.Marshal(expectedToken1)
						return tokenBytes, nil
					}

					secondNftKey := fmt.Sprintf("%s%s", "nft-0", string(big.NewInt(6).Bytes()))
					if strings.Contains(string(key), secondNftKey) {
						tokenBytes, _ := marshaller.Marshal(expectedToken2)
						return tokenBytes, nil
					}

					return nil, nil
				},
			}
			wrappedAccountMock := &state.AccountWrapMock{}
			wrappedAccountMock.SetTrackableDataTrie(&trieMock)

			return wrappedAccountMock, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.Pool{
		Txs: map[string]data.TransactionHandlerWithGasUsedAndFee{
			"hash0": outportcore.NewTransactionHandlerWithGasAndFee(&transaction.Transaction{
				SndAddr: []byte("addr"),
			}, 0, big.NewInt(0)),
		},
		Logs: []*data.LogData{
			{
				LogHandler: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("esdttoken"),
							},
						},
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								expectedToken1.TokenMetaData.Name,
								big.NewInt(0).SetUint64(expectedToken1.TokenMetaData.Nonce).Bytes(),
							},
						},
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								expectedToken2.TokenMetaData.Name,
								big.NewInt(0).SetUint64(expectedToken2.TokenMetaData.Nonce).Bytes(),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr := args.AddressConverter.Encode([]byte("addr"))
	require.Len(t, res, 1)
	require.Len(t, res[encodedAddr].Tokens, 3)

	require.Contains(t, res[encodedAddr].Tokens, &outportcore.AccountTokenData{
		Identifier: "esdttoken",
		Balance:    expectedToken0.Value.String(),
		Nonce:      0,
		MetaData:   nil,
	})

	require.Contains(t, res[encodedAddr].Tokens, &outportcore.AccountTokenData{
		Identifier: string(expectedToken1.TokenMetaData.Name),
		Balance:    expectedToken1.Value.String(),
		Nonce:      expectedToken1.TokenMetaData.Nonce,
		MetaData:   expectedToken1.TokenMetaData,
	})

	require.Contains(t, res[encodedAddr].Tokens, &outportcore.AccountTokenData{
		Identifier: string(expectedToken2.TokenMetaData.Name),
		Balance:    expectedToken2.Value.String(),
		Nonce:      expectedToken2.TokenMetaData.Nonce,
		MetaData:   expectedToken2.TokenMetaData,
	})

}

func getMockArgs() ArgsAlteredAccountsProvider {
	return ArgsAlteredAccountsProvider{
		ShardCoordinator:       &testscommon.ShardsCoordinatorMock{},
		AddressConverter:       &testscommon.PubkeyConverterMock{},
		AccountsDB:             &state.AccountsStub{},
		EsdtDataStorageHandler: &testscommon.EsdtStorageHandlerStub{},
	}
}
