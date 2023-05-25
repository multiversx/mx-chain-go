package alteredaccounts

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/alteredAccount"
	"github.com/multiversx/mx-chain-core-go/data/esdt"
	outportcore "github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/multiversx/mx-chain-core-go/data/rewardTx"
	"github.com/multiversx/mx-chain-core-go/data/smartContractResult"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/outport/process/alteredaccounts/shared"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/trie"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
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

func TestGetAlteredAccountFromUserAccount(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.AddressConverter = testscommon.NewPubkeyConverterMock(5)
	aap, _ := NewAlteredAccountsProvider(args)

	userAccount := &state.UserAccountStub{
		Balance:          big.NewInt(1000),
		DeveloperRewards: big.NewInt(100),
		Owner:            []byte("owner"),
		UserName:         []byte("contract"),
		Address:          []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	res := &alteredAccount.AlteredAccount{
		Address: "addr",
		Balance: "1000",
	}
	aap.addAdditionalDataInAlteredAccount(res, userAccount, &markedAlteredAccount{})

	require.Equal(t, &alteredAccount.AlteredAccount{
		Address: "addr",
		Balance: "1000",
		AdditionalData: &alteredAccount.AdditionalAccountData{
			DeveloperRewards: "100",
			CurrentOwner:     "6f776e6572",
			UserName:         "contract",
		},
	}, res)

	userAccount = &state.UserAccountStub{
		Balance:          big.NewInt(5000),
		DeveloperRewards: big.NewInt(5000),
		Owner:            []byte("own"),
		Address:          []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	}

	res = &alteredAccount.AlteredAccount{
		Address: "addr",
		Balance: "5000",
	}
	aap.addAdditionalDataInAlteredAccount(res, userAccount, &markedAlteredAccount{})

	require.Equal(t, &alteredAccount.AlteredAccount{
		Address: "addr",
		Balance: "5000",
		AdditionalData: &alteredAccount.AdditionalAccountData{
			DeveloperRewards: "5000",
		},
	}, res)
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
	t.Run("should not return balanceChanged for a receiver on an ESDTTransfer", testExtractAlteredAccountsFromPoolESDTTransferBalanceNotChanged)
	t.Run("should return balanceChanged for sender and receiver", testExtractAlteredAccountsFromPoolReceiverShouldHaveBalanceChanged)
	t.Run("should return balanceChanged only for sender", testExtractAlteredAccountsFromPoolOnlySenderShouldHaveBalanceChanged)
	t.Run("should return balanceChanged for sender nft create", textExtractAlteredAccountsFromPoolNftCreate)
	t.Run("should work with transaction value nil", textExtractAlteredAccountsFromPoolTransactionValueNil)
}

func testExtractAlteredAccountsFromPoolNoTransaction(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{}, shared.AlteredAccountsOptions{})
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender shard - tx0  "),
					RcvAddr: []byte("receiver shard - tx0"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash1": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender shard - tx1  "),
					RcvAddr: []byte("receiver shard - tx1"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key, info := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "sender"))
		require.True(t, info.AdditionalData.IsSender)
		require.True(t, info.AdditionalData.BalanceChanged)
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender shard - tx0  "),
					RcvAddr: []byte("receiver shard - tx0"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash1": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender shard - tx1  "),
					RcvAddr: []byte("receiver shard - tx1"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)
	require.Equal(t, 2, len(res))

	for key, info := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "receiver"))
		require.True(t, info.AdditionalData.BalanceChanged)
		require.False(t, info.AdditionalData.IsSender)
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("shard0 addr - tx0  "),
					RcvAddr: []byte("shard0 addr 2 - tx0"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash1": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("shard0 addr 3 - tx1"),
					RcvAddr: []byte("shard0 addr 3 - tx1"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash2": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("shard0 addr - tx2  "),
					RcvAddr: []byte("shard1 - tx2       "),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash3": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("shard1 addr - tx3  "),
					RcvAddr: []byte("shard0 addr - tx3  "),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"hash4": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("shard2 addr - tx4  "),
					RcvAddr: []byte("shard2 addr - tx3  "),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 5, len(res))

	for key := range res {
		decodedKey, _ := args.AddressConverter.Decode(key)
		require.True(t, strings.HasPrefix(string(decodedKey), "shard0"))
	}

	shard0AddrTx0, _ := args.AddressConverter.Encode([]byte("shard0 addr - tx0  "))
	require.Contains(t, res, shard0AddrTx0)
	shard0Addr2Tx0, _ := args.AddressConverter.Encode([]byte("shard0 addr 2 - tx0"))
	require.Contains(t, res, shard0Addr2Tx0)
	shard0Addr3Tx1, _ := args.AddressConverter.Encode([]byte("shard0 addr 3 - tx1"))
	require.Contains(t, res, shard0Addr3Tx1)
	shard0AddrTx2, _ := args.AddressConverter.Encode([]byte("shard0 addr - tx2  "))
	require.Contains(t, res, shard0AddrTx2)
	shard0AddrTx3, _ := args.AddressConverter.Encode([]byte("shard0 addr - tx3  "))
	require.Contains(t, res, shard0AddrTx3)
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender in shard 0  "),
					RcvAddr: []byte(receiverInSelfShard),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)
	require.Equal(t, 1, len(res))

	expectedAddressKey, err := args.AddressConverter.Encode([]byte(receiverInSelfShard))
	require.Nil(t, err)
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender in shard 0 - tx 0  "),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Rewards: map[string]*outportcore.RewardInfo{
			"hash1": {
				Reward: &rewardTx.RewardTx{
					RcvAddr: []byte("receiver in shard 0 - tx 1"),
					Value:   big.NewInt(1),
				},
			},
		},
		SmartContractResults: map[string]*outportcore.SCRInfo{
			"hash2": {
				SmartContractResult: &smartContractResult.SmartContractResult{
					SndAddr: []byte("sender in shard 0 - tx 2  "),
					RcvAddr: []byte("receiver in shard 0 - tx 2"),
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		InvalidTxs: map[string]*outportcore.TxInfo{
			"hash3": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("sender in shard 0 - tx 3  "),
					RcvAddr: []byte("receiver in shard 0 - tx 3"), // receiver for invalid txs should not be included
					Value:   big.NewInt(1),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
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
			return &state.StateUserAccountHandlerStub{}, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash",
				Log: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(0).Bytes(),
							},
						},
						{
							Address:    []byte("addr"), // other event for the same token, to ensure it isn't added twice
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(0).Bytes(),
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash",
				Log: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(0).Bytes(),
							},
						},
						{
							Address:    []byte("addr"), // other event for the same token, to ensure it isn't added twice
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"),
								big.NewInt(0).Bytes(),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr, _ := args.AddressConverter.Encode([]byte("addr"))
	require.Len(t, res, 1)
	require.Len(t, res[encodedAddr].Tokens, 1)
	require.Equal(t, &alteredAccount.AccountTokenData{
		Identifier: "token0",
		Balance:    expectedToken.Value.String(),
		Nonce:      0,
		Properties: "6f6b",
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash",
				Log: &transaction.Log{
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

	encodedAddr, _ := args.AddressConverter.Encode([]byte("addr"))
	require.Equal(t, &alteredAccount.AccountTokenData{
		Identifier: "token0",
		Balance:    expectedToken.Value.String(),
		Nonce:      expectedToken.TokenMetaData.Nonce,
		MetaData:   &alteredAccount.TokenMetaData{Nonce: expectedToken.TokenMetaData.Nonce},
	}, res[encodedAddr].Tokens[0])
}

func testExtractAlteredAccountsFromPoolShouldNotIncludeReceiverAddressIfNftCreateLog(t *testing.T) {
	t.Parallel()

	sendAddrShard0 := []byte("sender in shard 0 - tx 1  ")
	receiverOnDestination := []byte("receiver on destination shard")
	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce: 38,
		},
	}
	args := getMockArgs()
	args.AddressConverter = testscommon.NewPubkeyConverterMock(len(sendAddrShard0))
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hh": {
				Transaction: &transaction.Transaction{
					SndAddr: sendAddrShard0,
					RcvAddr: sendAddrShard0,
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				TxHash: "hh",
				Log: &transaction.Log{
					Address: sendAddrShard0,
					Events: []*transaction.Event{
						{
							Address:    sendAddrShard0,
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
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)

	sndAddrEncoded, err := args.AddressConverter.Encode(sendAddrShard0)
	require.Nil(t, err)
	require.Len(t, res, 1)
	require.True(t, res[sndAddrEncoded].Tokens[0].AdditionalData.IsNFTCreate)
	require.True(t, res[sndAddrEncoded].AdditionalData.BalanceChanged)
	require.True(t, res[sndAddrEncoded].AdditionalData.IsSender)

	mapKeyToSearch, err := args.AddressConverter.Encode(receiverOnDestination)
	require.Nil(t, err)
	require.Nil(t, res[mapKeyToSearch])
}

func testExtractAlteredAccountsFromPoolShouldIncludeDestinationFromTokensLogsTopics(t *testing.T) {
	t.Parallel()

	receiverOnDestination := []byte("receiver on destination shard")
	expectedToken := esdt.ESDigitalToken{
		Value: big.NewInt(37),
		TokenMetaData: &esdt.MetaData{
			Nonce:      38,
			Name:       []byte("name"),
			Creator:    []byte("creator"),
			Royalties:  1000,
			Hash:       []byte("hash"),
			URIs:       [][]byte{[]byte("uri1"), []byte("uri2")},
			Attributes: []byte("attributes"),
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash0",
				Log: &transaction.Log{
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

	mapKeyToSearch, err := args.AddressConverter.Encode(receiverOnDestination)
	require.Nil(t, err)
	creator, err := args.AddressConverter.Encode(expectedToken.TokenMetaData.Creator)
	require.Nil(t, err)
	require.Len(t, res[mapKeyToSearch].Tokens, 1)
	require.Equal(t, res[mapKeyToSearch].Tokens[0], &alteredAccount.AccountTokenData{
		Identifier: "token0",
		Balance:    "37",
		Nonce:      38,
		MetaData: &alteredAccount.TokenMetaData{
			Nonce:      38,
			Name:       "name",
			Creator:    creator,
			Royalties:  1000,
			Hash:       []byte("hash"),
			URIs:       [][]byte{[]byte("uri1"), []byte("uri2")},
			Attributes: []byte("attributes"),
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

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("addr"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash0",
				Log: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("esdt"),
								big.NewInt(1).Bytes(),
							},
						},
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTTransfer),
							Topics: [][]byte{
								[]byte("nft"),
								big.NewInt(38).Bytes(),
								big.NewInt(1).Bytes(),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{})
	require.NoError(t, err)

	encodedAddr, _ := args.AddressConverter.Encode([]byte("addr"))
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
				RetrieveValueCalled: func(key []byte) ([]byte, uint32, error) {
					if strings.Contains(string(key), "esdttoken") {
						tokenBytes, _ := marshaller.Marshal(expectedToken0)
						return tokenBytes, 0, nil
					}

					firstNftKey := fmt.Sprintf("%s%s", "nft-0", string(big.NewInt(5).Bytes()))
					if strings.Contains(string(key), firstNftKey) {
						tokenBytes, _ := marshaller.Marshal(expectedToken1)
						return tokenBytes, 0, nil
					}

					secondNftKey := fmt.Sprintf("%s%s", "nft-0", string(big.NewInt(6).Bytes()))
					if strings.Contains(string(key), secondNftKey) {
						tokenBytes, _ := marshaller.Marshal(expectedToken2)
						return tokenBytes, 0, nil
					}

					return nil, 0, nil
				},
			}
			wrappedAccountMock := &state.AccountWrapMock{}
			wrappedAccountMock.SetTrackableDataTrie(&trieMock)

			return wrappedAccountMock, nil
		},
	}
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"hash0": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("addr"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				TxHash: "hash0",
				Log: &transaction.Log{
					Address: []byte("addr"),
					Events: []*transaction.Event{
						{
							Address:    []byte("addr"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("esdttoken"),
								big.NewInt(0).Bytes(),
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

	encodedAddr, _ := args.AddressConverter.Encode([]byte("addr"))
	require.Len(t, res, 1)
	require.Len(t, res[encodedAddr].Tokens, 3)

	require.Contains(t, res[encodedAddr].Tokens, &alteredAccount.AccountTokenData{
		Identifier: "esdttoken",
		Balance:    expectedToken0.Value.String(),
		Nonce:      0,
		MetaData:   nil,
	})

	require.Contains(t, res[encodedAddr].Tokens, &alteredAccount.AccountTokenData{
		Identifier: string(expectedToken1.TokenMetaData.Name),
		Balance:    expectedToken1.Value.String(),
		Nonce:      expectedToken1.TokenMetaData.Nonce,
		MetaData: &alteredAccount.TokenMetaData{
			Nonce: expectedToken1.TokenMetaData.Nonce,
			Name:  string(expectedToken1.TokenMetaData.Name),
		},
	})

	require.Contains(t, res[encodedAddr].Tokens, &alteredAccount.AccountTokenData{
		Identifier: string(expectedToken2.TokenMetaData.Name),
		Balance:    expectedToken2.Value.String(),
		Nonce:      expectedToken2.TokenMetaData.Nonce,
		MetaData: &alteredAccount.TokenMetaData{
			Nonce: expectedToken2.TokenMetaData.Nonce,
			Name:  string(expectedToken2.TokenMetaData.Name),
		},
	})

}

func testExtractAlteredAccountsFromPoolESDTTransferBalanceNotChanged(t *testing.T) {
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
			return &state.AccountWrapMock{
				Balance: big.NewInt(10),
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(3)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"txHash": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("rcv"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				TxHash: "txHash",
				Log: &transaction.Log{
					Address: []byte("snd"),
					Events: []*transaction.Event{
						{
							Address:    []byte("snd"),
							Identifier: []byte(core.BuiltInFunctionESDTTransfer),
							Topics: [][]byte{
								[]byte("token0"), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), []byte("rcv"),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)

	encodedAddrSnd, _ := args.AddressConverter.Encode([]byte("snd"))
	encodedAddrRcv, _ := args.AddressConverter.Encode([]byte("rcv"))
	require.Equal(t, map[string]*alteredAccount.AlteredAccount{
		encodedAddrSnd: {
			Address: encodedAddrSnd,
			Balance: "10",
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "token0",
					Balance:    expectedToken.Value.String(),
					Nonce:      0,
					Properties: "6f6b",
					MetaData:   nil,
					AdditionalData: &alteredAccount.AdditionalAccountTokenData{
						IsNFTCreate: false,
					},
				},
			},
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
				IsSender:       true,
			},
		},
		encodedAddrRcv: {
			Address: encodedAddrRcv,
			Balance: "10",
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "token0",
					Balance:    expectedToken.Value.String(),
					Nonce:      0,
					Properties: "6f6b",
					AdditionalData: &alteredAccount.AdditionalAccountTokenData{
						IsNFTCreate: false,
					},
				},
			},
			AdditionalData: &alteredAccount.AdditionalAccountData{
				IsSender:       false,
				BalanceChanged: false,
			},
		},
	}, res)
}

func testExtractAlteredAccountsFromPoolReceiverShouldHaveBalanceChanged(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{
				Balance: big.NewInt(15),
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(3)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"txHash": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("rcv"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
			"txHash2": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("rcv"),
					Value:   big.NewInt(2),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})

	require.NoError(t, err)

	encodedAddrSnd, _ := args.AddressConverter.Encode([]byte("snd"))
	encodedAddrRcv, _ := args.AddressConverter.Encode([]byte("rcv"))
	require.Equal(t, map[string]*alteredAccount.AlteredAccount{
		encodedAddrSnd: {
			Address: encodedAddrSnd,
			Balance: "15",
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
				IsSender:       true,
			},
		},
		encodedAddrRcv: {
			Address: encodedAddrRcv,
			Balance: "15",
			AdditionalData: &alteredAccount.AdditionalAccountData{
				IsSender:       false,
				BalanceChanged: true,
			},
		},
	}, res)
}

func testExtractAlteredAccountsFromPoolOnlySenderShouldHaveBalanceChanged(t *testing.T) {
	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{
				Balance: big.NewInt(15),
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(3)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"txHash": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("rcv"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)

	encodedAddrSnd, _ := args.AddressConverter.Encode([]byte("snd"))
	require.Equal(t, map[string]*alteredAccount.AlteredAccount{
		encodedAddrSnd: {
			Address: encodedAddrSnd,
			Balance: "15",
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
				IsSender:       true,
			},
		},
	}, res)
}

func textExtractAlteredAccountsFromPoolNftCreate(t *testing.T) {
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
			return &state.AccountWrapMock{
				Balance: big.NewInt(10),
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(3)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"txHash": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("snd"),
					Value:   big.NewInt(0),
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
		Logs: []*outportcore.LogData{
			{
				TxHash: "txHash",
				Log: &transaction.Log{
					Address: []byte("snd"),
					Events: []*transaction.Event{
						{
							Address:    []byte("snd"),
							Identifier: []byte(core.BuiltInFunctionESDTNFTCreate),
							Topics: [][]byte{
								[]byte("token0"), big.NewInt(0).Bytes(), big.NewInt(10).Bytes(), []byte("a"),
							},
						},
					},
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})
	require.NoError(t, err)

	encodedAddrSnd, _ := args.AddressConverter.Encode([]byte("snd"))
	require.Equal(t, map[string]*alteredAccount.AlteredAccount{
		encodedAddrSnd: {
			Address: encodedAddrSnd,
			Balance: "10",
			Tokens: []*alteredAccount.AccountTokenData{
				{
					Identifier: "token0",
					Balance:    expectedToken.Value.String(),
					Nonce:      0,
					Properties: "6f6b",
					MetaData:   nil,
					AdditionalData: &alteredAccount.AdditionalAccountTokenData{
						IsNFTCreate: true,
					},
				},
			},
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
				IsSender:       true,
			},
		},
	}, res)
}

func textExtractAlteredAccountsFromPoolTransactionValueNil(t *testing.T) {
	t.Parallel()

	args := getMockArgs()
	args.AccountsDB = &state.AccountsStub{
		LoadAccountCalled: func(_ []byte) (vmcommon.AccountHandler, error) {
			return &state.AccountWrapMock{
				Balance: big.NewInt(15),
			}, nil
		},
	}
	args.AddressConverter = testscommon.NewPubkeyConverterMock(3)
	aap, _ := NewAlteredAccountsProvider(args)

	res, err := aap.ExtractAlteredAccountsFromPool(&outportcore.TransactionPool{
		Transactions: map[string]*outportcore.TxInfo{
			"txHash": {
				Transaction: &transaction.Transaction{
					SndAddr: []byte("snd"),
					RcvAddr: []byte("rcv"),
					Value:   nil,
				},
				FeeInfo: &outportcore.FeeInfo{
					Fee: big.NewInt(0),
				},
			},
		},
	}, shared.AlteredAccountsOptions{
		WithAdditionalOutportData: true,
	})

	require.NoError(t, err)

	encodedAddrSnd, _ := args.AddressConverter.Encode([]byte("snd"))
	require.Equal(t, map[string]*alteredAccount.AlteredAccount{
		encodedAddrSnd: {
			Address: encodedAddrSnd,
			Balance: "15",
			AdditionalData: &alteredAccount.AdditionalAccountData{
				BalanceChanged: true,
				IsSender:       true,
			},
		},
	}, res)
}

func getMockArgs() ArgsAlteredAccountsProvider {
	return ArgsAlteredAccountsProvider{
		ShardCoordinator:       &testscommon.ShardsCoordinatorMock{},
		AddressConverter:       &testscommon.PubkeyConverterMock{},
		AccountsDB:             &state.AccountsStub{},
		EsdtDataStorageHandler: &testscommon.EsdtStorageHandlerStub{},
	}
}
