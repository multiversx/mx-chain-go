package external_test

import (
	"context"
	"encoding/hex"
	"errors"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding/nodesCoordinator"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createMockArgs() external.ArgNodeApiResolver {
	return external.ArgNodeApiResolver{
		SCQueryService:           &mock.SCQueryServiceStub{},
		StatusMetricsHandler:     &testscommon.StatusMetricsStub{},
		TxCostHandler:            &mock.TransactionCostEstimatorMock{},
		TotalStakedValueHandler:  &mock.StakeValuesProcessorStub{},
		DirectStakedListHandler:  &mock.DirectStakedListProcessorStub{},
		DelegatedListHandler:     &mock.DelegatedListProcessorStub{},
		APIBlockHandler:          &mock.BlockAPIHandlerStub{},
		APITransactionHandler:    &mock.TransactionAPIHandlerStub{},
		APIInternalBlockHandler:  &mock.InternalBlockApiHandlerStub{},
		GenesisNodesSetupHandler: &testscommon.NodesSetupStub{},
		ValidatorPubKeyConverter: &testscommon.PubkeyConverterMock{},
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
	arg.TxCostHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTransactionCostHandler, err)
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

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArgs()
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(nar))
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
		ExecuteQueryCalled: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)

	_, _ = nar.ExecuteSCQuery(&process.SCQuery{
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
			GetBlockByNonceCalled: func(nonce uint64, withTxs bool) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByNonce(10, true)
		require.True(t, wasCalled)
	})

	t.Run("GetBlockByHash", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		arg := createMockArgs()
		arg.APIBlockHandler = &mock.BlockAPIHandlerStub{
			GetBlockByHashCalled: func(hash []byte, withTxs bool) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByHash("0101", true)
		require.True(t, wasCalled)
	})

	t.Run("GetBlockByRound", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		arg := createMockArgs()
		arg.APIBlockHandler = &mock.BlockAPIHandlerStub{
			GetBlockByRoundCalled: func(round uint64, withTxs bool) (*api.Block, error) {
				wasCalled = true
				return nil, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)

		_, _ = nar.GetBlockByRound(10, true)
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
			GetTransactionsPoolCalled: func() (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPool()
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		expectedTxsPool := &common.TransactionsPoolAPIResponse{
			RegularTransactions:  []string{"txhash1"},
			SmartContractResults: []string{"txhash2"},
			Rewards:              []string{"txhash3"},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolCalled: func() (*common.TransactionsPoolAPIResponse, error) {
				return expectedTxsPool, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPool()
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
		providedParameters := "sender,hash,receiver"
		expectedResponse := &common.TransactionsPoolForSenderApiResponse{
			Transactions: []common.Transaction{
				{
					Hash:     "txHash1",
					Sender:   expectedSender,
					Receiver: "receiver1",
				},
				{
					Hash:     "txHash2",
					Sender:   expectedSender,
					Receiver: "receiver2",
				},
			},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolForSenderCalled: func(sender, parameters string) (*common.TransactionsPoolForSenderApiResponse, error) {
				require.Equal(t, expectedSender, sender)
				require.Equal(t, providedParameters, parameters)
				return expectedResponse, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolForSender(expectedSender, providedParameters)
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
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolNonceGapsForSender("sender")
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
					From: "33",
					To:   "60",
				},
			},
		}
		arg := createMockArgs()
		arg.APITransactionHandler = &mock.TransactionAPIHandlerStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return expectedNonceGaps, nil
			},
		}

		nar, _ := external.NewNodeApiResolver(arg)
		res, err := nar.GetTransactionsPoolNonceGapsForSender(expectedSender)
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
