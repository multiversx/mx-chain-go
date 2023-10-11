package external_test

import (
	"context"
	"encoding/hex"
	"errors"
	"math/big"
	"testing"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/data/api"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/genesis"
	"github.com/multiversx/mx-chain-go/genesis/data"
	"github.com/multiversx/mx-chain-go/node/external"
	"github.com/multiversx/mx-chain-go/node/mock"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/genesisMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	vmcommon "github.com/multiversx/mx-chain-vm-common-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var expectedErr = errors.New("expected error")

func createMockArgs() external.ArgNodeApiResolver {
	return external.ArgNodeApiResolver{
		SCQueryService:           &mock.SCQueryServiceStub{},
		StatusMetricsHandler:     &testscommon.StatusMetricsStub{},
		APITransactionEvaluator:  &mock.TransactionCostEstimatorMock{},
		TotalStakedValueHandler:  &mock.StakeValuesProcessorStub{},
		DirectStakedListHandler:  &mock.DirectStakedListProcessorStub{},
		DelegatedListHandler:     &mock.DelegatedListProcessorStub{},
		APIBlockHandler:          &mock.BlockAPIHandlerStub{},
		APITransactionHandler:    &mock.TransactionAPIHandlerStub{},
		APIInternalBlockHandler:  &mock.InternalBlockApiHandlerStub{},
		GenesisNodesSetupHandler: &testscommon.NodesSetupStub{},
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
		AccountsParser:           &genesisMocks.AccountsParserStub{},
		GasScheduleNotifier:      &testscommon.GasScheduleNotifierMock{},
		ManagedPeersMonitor:      &testscommon.ManagedPeersMonitorStub{},
	}
}

func TestNewNodeApiResolver_NilSCQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.SCQueryService = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilSCQueryService, err)
}

func TestNewNodeApiResolver_NilStatusMetricsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.StatusMetricsHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilStatusMetrics, err)
}

func TestNewNodeApiResolver_NilTransactionCostEstimator(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.APITransactionEvaluator = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilAPITransactionEvaluator, err)
}

func TestNewNodeApiResolver_NilTotalStakedValueHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.TotalStakedValueHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTotalStakedValueHandler, err)
}

func TestNewNodeApiResolver_NilDirectStakedListHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.DirectStakedListHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilDirectStakeListHandler, err)
}

func TestNewNodeApiResolver_NilDelegatedListHandler(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.DelegatedListHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilDelegatedListHandler, err)
}

func TestNewNodeApiResolver_NilGasSchedules(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	arg.GasScheduleNotifier = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilGasScheduler, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, err)
	assert.NotNil(t, nar)
}

func TestNodeApiResolver_CloseShouldReturnNil(t *testing.T) {
	t.Parallel()

	args := createMockArgs()
	closeCalled := false
	args.SCQueryService = &mock.SCQueryServiceStub{
		CloseCalled: func() error {
			closeCalled = true

			return nil
		},
	}
	nar, _ := external.NewNodeApiResolver(args)

	err := nar.Close()
	assert.Nil(t, err)
	assert.True(t, closeCalled)
}

func TestNodeApiResolver_GetDataValueShouldCall(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	wasCalled := false
	arg.SCQueryService = &mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, info common.BlockInfo, e error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, info, nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)

	_, _, _ = nar.ExecuteSCQuery(&process.SCQuery{
		ScAddress: []byte{0},
		FuncName:  "",
	})

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapWithoutP2PShouldBeCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	wasCalled := false
	arg.StatusMetricsHandler = &testscommon.StatusMetricsStub{
		StatusMetricsMapWithoutP2PCalled: func() (map[string]interface{}, error) {
			wasCalled = true
			return nil, nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_, _ = nar.StatusMetrics().StatusMetricsMapWithoutP2P()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusP2PMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	wasCalled := false
	arg.StatusMetricsHandler = &testscommon.StatusMetricsStub{
		StatusP2pMetricsMapCalled: func() (map[string]interface{}, error) {
			wasCalled = true
			return nil, nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_, _ = nar.StatusMetrics().StatusP2pMetricsMap()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_NetworkMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	wasCalled := false
	arg.StatusMetricsHandler = &testscommon.StatusMetricsStub{
		NetworkMetricsCalled: func() (map[string]interface{}, error) {
			wasCalled = true
			return nil, nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_, _ = nar.StatusMetrics().NetworkMetrics()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_GetTotalStakedValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArgs()
	stakeValue := &api.StakeValues{}
	arg.TotalStakedValueHandler = &mock.StakeValuesProcessorStub{
		GetTotalStakedValueCalled: func(_ context.Context) (*api.StakeValues, error) {
			wasCalled = true
			return stakeValue, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredStakeValue, err := nar.GetTotalStakedValue(context.Background())
	assert.Nil(t, err)
	assert.True(t, recoveredStakeValue == stakeValue) //pointer testing
	assert.True(t, wasCalled)
}

func TestNodeApiResolver_GetDelegatorsList(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArgs()
	delegators := make([]*api.Delegator, 1)
	arg.DelegatedListHandler = &mock.DelegatedListProcessorStub{
		GetDelegatorsListCalled: func(_ context.Context) ([]*api.Delegator, error) {
			wasCalled = true
			return delegators, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredDelegatorsList, err := nar.GetDelegatorsList(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, recoveredDelegatorsList, delegators)
	assert.True(t, wasCalled)
}

func TestNodeApiResolver_GetDirectStakedList(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArgs()
	directStakedValueList := make([]*api.DirectStakedValue, 1)
	arg.DirectStakedListHandler = &mock.DirectStakedListProcessorStub{
		GetDirectStakedListCalled: func(ctx context.Context) ([]*api.DirectStakedValue, error) {
			wasCalled = true
			return directStakedValueList, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredDirectStakedValueList, err := nar.GetDirectStakedList(context.Background())
	assert.Nil(t, err)
	assert.Equal(t, recoveredDirectStakedValueList, directStakedValueList)
	assert.True(t, wasCalled)
}

func TestNodeApiResolver_APIBlockHandler(t *testing.T) {
	t.Parallel()

	t.Run("GetBlockByNonce", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		arg := createMockArgs()
		arg.APIBlockHandler = &mock.BlockAPIHandlerStub{
			GetBlockByNonceCalled: func(nonce uint64, options api.BlockQueryOptions) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByNonce(10, api.BlockQueryOptions{WithTransactions: true})
		require.True(t, wasCalled)
	})

	t.Run("GetBlockByHash", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		arg := createMockArgs()
		arg.APIBlockHandler = &mock.BlockAPIHandlerStub{
			GetBlockByHashCalled: func(hash []byte, options api.BlockQueryOptions) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByHash("0101", api.BlockQueryOptions{WithTransactions: true})
		require.True(t, wasCalled)
	})

	t.Run("GetBlockByRound", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		arg := createMockArgs()
		arg.APIBlockHandler = &mock.BlockAPIHandlerStub{
			GetBlockByRoundCalled: func(round uint64, options api.BlockQueryOptions) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByRound(10, api.BlockQueryOptions{WithTransactions: true})
		require.True(t, wasCalled)
	})
}

func TestNodeApiResolver_APITransactionHandler(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArgs()
	arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
		GetTransactionCalled: func(hash string, withResults bool) (*transaction.ApiTransactionResult, error) {
			wasCalled = true
			return nil, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)

	_, _ = nar.GetTransaction("0101", true)
	require.True(t, wasCalled)
}

func TestNodeApiResolver_GetTransactionsPool(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPool("")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTxsPool := &common.TransactionsPoolAPIResponse{
			RegularTransactions: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "txhash1",
					},
				},
			},
			SmartContractResults: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "txhash2",
					},
				},
			},
			Rewards: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash": "txhash3",
					},
				},
			},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolCalled: func(fields string) (*common.TransactionsPoolAPIResponse, error) {
				return expectedTxsPool, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPool("")
		require.NoError(t, err)
		require.Equal(t, expectedTxsPool, res)
	})
}

func TestNodeApiResolver_GetTransactionsPoolForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolForSenderCalled: func(sender, parameters string) (*common.TransactionsPoolForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolForSender("sender", "")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSender := "alice"
		providedFields := "sender,receiver"
		expectedResponse := &common.TransactionsPoolForSenderApiResponse{
			Transactions: []common.Transaction{
				{
					TxFields: map[string]interface{}{
						"hash":     "txhash1",
						"sender":   expectedSender,
						"receiver": "receiver1",
					},
				},
				{
					TxFields: map[string]interface{}{
						"hash":     "txhash2",
						"sender":   expectedSender,
						"receiver": "receiver2",
					},
				},
			},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolForSenderCalled: func(sender, fields string) (*common.TransactionsPoolForSenderApiResponse, error) {
				require.Equal(t, expectedSender, sender)
				require.Equal(t, providedFields, fields)
				return expectedResponse, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolForSender(expectedSender, providedFields)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, res)
	})
}

func TestNodeApiResolver_GetLastPoolNonceForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return 0, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetLastPoolNonceForSender("sender")
		require.Equal(t, uint64(0), res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSender := "alice"
		expectedNonce := uint64(33)
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return expectedNonce, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetLastPoolNonceForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedNonce, res)
	})
}

func TestNodeApiResolver_GetTransactionsPoolNonceGapsForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected error")
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolNonceGapsForSender("sender", 1)
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedSender := "alice"
		expectedNonceGaps := &common.TransactionsPoolNonceGapsForSenderApiResponse{
			Sender: expectedSender,
			Gaps: []common.NonceGapApiResponse{
				{
					From: 33,
					To:   60,
				},
			},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string, senderAccountNonce uint64) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return expectedNonceGaps, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolNonceGapsForSender(expectedSender, 0)
		require.NoError(t, err)
		require.Equal(t, expectedNonceGaps, res)
	})
}

func TestNodeApiResolver_GetGenesisNodesPubKeys(t *testing.T) {
	t.Parallel()

	pubKey1 := []byte("pubKey1")
	expPubKey1 := hex.EncodeToString(pubKey1)
	pubKey2 := []byte("pubKey2")
	expPubKey2 := hex.EncodeToString(pubKey2)

	eligible := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
		1: {shardingMocks.NewNodeInfo([]byte("address1"), pubKey1, 1, 1)},
	}
	waiting := map[uint32][]nodesCoordinator.GenesisNodeInfoHandler{
		1: {shardingMocks.NewNodeInfo([]byte("address2"), pubKey2, 1, 1)},
	}

	arg := createMockArgs()
	arg.GenesisNodesSetupHandler = &testscommon.NodesSetupStub{
		InitialNodesInfoCalled: func() (map[uint32][]nodesCoordinator.GenesisNodeInfoHandler, map[uint32][]nodesCoordinator.GenesisNodeInfoHandler) {
			return eligible, waiting
		},
	}

	nar, err := external.NewNodeApiResolver(arg)
	require.Nil(t, err)

	expectedEligible := map[uint32][]string{
		1: {expPubKey1},
	}
	expectedWaiting := map[uint32][]string{
		1: {expPubKey2},
	}

	el, wt := nar.GetGenesisNodesPubKeys()
	assert.Equal(t, expectedEligible, el)
	assert.Equal(t, expectedWaiting, wt)
}

func TestNodeApiResolver_GetGenesisBalances(t *testing.T) {
	t.Parallel()

	t.Run("should return empty slice", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.AccountsParser = &genesisMocks.AccountsParserStub{
			InitialAccountsCalled: func() []genesis.InitialAccountHandler {
				return nil
			},
		}

		nar, _ := external.NewNodeApiResolver(args)

		res, err := nar.GetGenesisBalances()
		require.NoError(t, err)
		require.Empty(t, res)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		initialAccount := &data.InitialAccount{
			Address:      "addr",
			Supply:       big.NewInt(100),
			Balance:      big.NewInt(110),
			StakingValue: big.NewInt(120),
			Delegation: &data.DelegationData{
				Address: "addr2",
				Value:   big.NewInt(130),
			},
		}
		args := createMockArgs()
		args.AccountsParser = &genesisMocks.AccountsParserStub{
			InitialAccountsCalled: func() []genesis.InitialAccountHandler {
				return []genesis.InitialAccountHandler{initialAccount}
			},
		}

		nar, _ := external.NewNodeApiResolver(args)

		res, err := nar.GetGenesisBalances()
		require.NoError(t, err)
		require.Equal(t, []*common.InitialAccountAPI{
			{
				Address:      initialAccount.Address,
				Supply:       "100",
				Balance:      "110",
				StakingValue: "120",
				Delegation: common.DelegationDataAPI{
					Address: initialAccount.Delegation.Address,
					Value:   "130",
				},
			},
		}, res)
	})
}

func TestNodeApiResolver_GetGasConfigs(t *testing.T) {
	t.Parallel()

	args := createMockArgs()

	wasCalled := false
	args.GasScheduleNotifier = &testscommon.GasScheduleNotifierMock{
		LatestGasScheduleCopyCalled: func() map[string]map[string]uint64 {
			wasCalled = true
			return nil
		},
	}

	nar, err := external.NewNodeApiResolver(args)
	require.Nil(t, err)

	_ = nar.GetGasConfigs()
	require.True(t, wasCalled)
}

func TestNodeApiResolver_GetManagedKeysCount(t *testing.T) {
	t.Parallel()

	providedCount := 100
	args := createMockArgs()
	args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
		GetManagedKeysCountCalled: func() int {
			return providedCount
		},
	}
	nar, err := external.NewNodeApiResolver(args)
	require.NoError(t, err)

	count := nar.GetManagedKeysCount()
	require.Equal(t, providedCount, count)
}

func TestNodeApiResolver_GetManagedKeys(t *testing.T) {
	t.Parallel()

	providedKeys := [][]byte{
		[]byte("pk1"),
		[]byte("pk2"),
	}
	expectedKeys := []string{
		"pk1",
		"pk2",
	}
	args := createMockArgs()
	args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
		GetManagedKeysCalled: func() [][]byte {
			return providedKeys
		},
	}
	args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
		SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
			return string(pkBytes)
		},
	}
	nar, err := external.NewNodeApiResolver(args)
	require.NoError(t, err)

	keys := nar.GetManagedKeys()
	require.Equal(t, expectedKeys, keys)
}

func TestNodeApiResolver_GetEligibleManagedKeys(t *testing.T) {
	t.Parallel()

	t.Run("monitor error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
			GetEligibleManagedKeysCalled: func() ([][]byte, error) {
				return nil, expectedErr
			},
		}
		nar, err := external.NewNodeApiResolver(args)
		require.NoError(t, err)

		keys, err := nar.GetEligibleManagedKeys()
		require.Equal(t, expectedErr, err)
		require.Nil(t, keys)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedKeys := [][]byte{
			[]byte("key1"),
			[]byte("key2"),
		}
		expectedKeys := []string{
			"key1",
			"key2",
		}
		args := createMockArgs()
		args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
			GetEligibleManagedKeysCalled: func() ([][]byte, error) {
				return providedKeys, nil
			},
		}
		args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
			SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
				return string(pkBytes)
			},
		}
		nar, err := external.NewNodeApiResolver(args)
		require.NoError(t, err)

		keys, err := nar.GetEligibleManagedKeys()
		require.NoError(t, err)
		require.Equal(t, expectedKeys, keys)
	})
}

func TestNodeApiResolver_GetWaitingManagedKeys(t *testing.T) {
	t.Parallel()

	t.Run("monitor error should error", func(t *testing.T) {
		t.Parallel()

		args := createMockArgs()
		args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
			GetWaitingManagedKeysCalled: func() ([][]byte, error) {
				return nil, expectedErr
			},
		}
		nar, err := external.NewNodeApiResolver(args)
		require.NoError(t, err)

		keys, err := nar.GetWaitingManagedKeys()
		require.Equal(t, expectedErr, err)
		require.Nil(t, keys)
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		providedKeys := [][]byte{
			[]byte("key1"),
			[]byte("key2"),
		}
		expectedKeys := []string{
			"key1",
			"key2",
		}
		args := createMockArgs()
		args.ManagedPeersMonitor = &testscommon.ManagedPeersMonitorStub{
			GetWaitingManagedKeysCalled: func() ([][]byte, error) {
				return providedKeys, nil
			},
		}
		args.ValidatorPubKeyConverter = &testscommon.PubkeyConverterStub{
			SilentEncodeCalled: func(pkBytes []byte, log core.Logger) string {
				return string(pkBytes)
			},
		}
		nar, err := external.NewNodeApiResolver(args)
		require.NoError(t, err)

		keys, err := nar.GetWaitingManagedKeys()
		require.NoError(t, err)
		require.Equal(t, expectedKeys, keys)
	})
}

func TestNodeApiResolver_IsInterfaceNil(t *testing.T) {
	t.Parallel()

	nar, _ := external.NewNodeApiResolver(external.ArgNodeApiResolver{})
	require.True(t, nar.IsInterfaceNil())

	arg := createMockArgs()
	nar, _ = external.NewNodeApiResolver(arg)
	require.False(t, nar.IsInterfaceNil())
}
