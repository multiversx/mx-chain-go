package facade

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	atomicCore "github.com/ElrondNetwork/elrond-go-core/core/atomic"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	nodeData "github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/api"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-core/data/esdt"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/data/vm"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/debug"
	"github.com/ElrondNetwork/elrond-go/facade/mock"
	"github.com/ElrondNetwork/elrond-go/heartbeat/data"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO increase code coverage

func createMockArguments() ArgNodeFacade {
	return ArgNodeFacade{
		Node:                   &mock.NodeStub{},
		ApiResolver:            &mock.ApiResolverStub{},
		RestAPIServerDebugMode: false,
		TxSimulatorProcessor:   &mock.TxExecutionSimulatorStub{},
		WsAntifloodConfig: config.WebServerAntifloodConfig{
			SimultaneousRequests:               1,
			SameSourceRequests:                 1,
			SameSourceResetIntervalInSec:       1,
			TrieOperationsDeadlineMilliseconds: 1,
		},
		FacadeConfig: config.FacadeConfig{
			RestApiInterface: "127.0.0.1:8080",
			PprofEnabled:     false,
		},
		ApiRoutesConfig: config.ApiRoutesConfig{APIPackages: map[string]config.APIPackageConfig{
			"node": {
				Routes: []config.RouteConfig{
					{Name: "status"},
				},
			},
		}},
		AccountsState: &stateMock.AccountsStub{},
		PeerState:     &stateMock.AccountsStub{},
		Blockchain: &testscommon.ChainHandlerStub{
			GetCurrentBlockHeaderCalled: func() nodeData.HeaderHandler {
				return &block.Header{}
			},
			GetCurrentBlockRootHashCalled: func() []byte {
				return []byte("root hash")
			},
		},
	}
}

// ------- NewNodeFacade

func TestNewNodeFacade_WithNilNodeShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Node = nil
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.Equal(t, ErrNilNode, err)
}

func TestNewNodeFacade_WithNilApiResolverShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ApiResolver = nil
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.Equal(t, ErrNilApiResolver, err)
}

func TestNewNodeFacade_WithInvalidSimultaneousRequestsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SimultaneousRequests = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidSameSourceResetIntervalInSecShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SameSourceResetIntervalInSec = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidSameSourceRequestsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.SameSourceRequests = 0
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrInvalidValue))
}

func TestNewNodeFacade_WithInvalidApiRoutesConfigShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.ApiRoutesConfig = config.ApiRoutesConfig{}
	nf, err := NewNodeFacade(arg)

	assert.True(t, check.IfNil(nf))
	assert.True(t, errors.Is(err, ErrNoApiRoutesConfig))
}

func TestNewNodeFacade_WithValidNodeShouldReturnNotNil(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, err := NewNodeFacade(arg)

	assert.False(t, check.IfNil(nf))
	assert.Nil(t, err)
}

// ------- Methods

func TestNodeFacade_GetBalanceWithValidAddressShouldReturnBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(addr)

	assert.Nil(t, err)
	assert.Equal(t, balance, amount)
}

func TestNodeFacade_GetBalanceWithUnknownAddressShouldReturnZeroBalance(t *testing.T) {
	t.Parallel()

	balance := big.NewInt(10)
	addr := "testAddress"
	unknownAddr := "unknownAddr"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			if addr == address {
				return balance, nil
			}
			return big.NewInt(0), nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(unknownAddr)
	assert.Nil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetBalanceWithErrorOnNodeShouldReturnZeroBalanceAndError(t *testing.T) {
	t.Parallel()

	addr := "testAddress"
	zeroBalance := big.NewInt(0)

	node := &mock.NodeStub{
		GetBalanceHandler: func(address string) (*big.Int, error) {
			return big.NewInt(0), errors.New("error on getBalance on node")
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	amount, err := nf.GetBalance(addr)
	assert.NotNil(t, err)
	assert.Equal(t, zeroBalance, amount)
}

func TestNodeFacade_GetTransactionWithValidInputsShouldNotReturnError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	node := &mock.NodeStub{}

	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction(testHash, false)
	assert.Nil(t, err)
	assert.Equal(t, testTx, tx)
}

func TestNodeFacade_GetTransactionWithUnknowHashShouldReturnNilAndNoError(t *testing.T) {
	t.Parallel()

	testHash := "testHash"
	testTx := &transaction.ApiTransactionResult{}
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTransactionHandler: func(hash string, withEvents bool) (*transaction.ApiTransactionResult, error) {
			if hash == testHash {
				return testTx, nil
			}
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	tx, err := nf.GetTransaction("unknownHash", false)
	assert.Nil(t, err)
	assert.Nil(t, tx)
}

func TestNodeFacade_SetSyncer(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	sync := &mock.SyncTimerMock{}
	nf.SetSyncer(sync)
	assert.Equal(t, sync, nf.GetSyncer())
}

func TestNodeFacade_GetAccount(t *testing.T) {
	t.Parallel()

	getAccountCalled := false
	node := &mock.NodeStub{}
	node.GetAccountHandler = func(address string) (api.AccountResponse, error) {
		getAccountCalled = true
		return api.AccountResponse{}, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	_, _ = nf.GetAccount("test")
	assert.True(t, getAccountCalled)
}

func TestNodeFacade_GetUsername(t *testing.T) {
	t.Parallel()

	expectedUsername := "username"
	node := &mock.NodeStub{}
	node.GetUsernameCalled = func(address string) (string, error) {
		return expectedUsername, nil
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	username, err := nf.GetUsername("test")
	assert.NoError(t, err)
	assert.Equal(t, expectedUsername, username)
}

func TestNodeFacade_GetHeartbeatsReturnsNilShouldErr(t *testing.T) {
	t.Parallel()

	node := &mock.NodeStub{
		GetHeartbeatsHandler: func() []data.PubKeyHeartbeat {
			return nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	result, err := nf.GetHeartbeats()

	assert.Nil(t, result)
	assert.Equal(t, ErrHeartbeatsNotActive, err)
}

func TestNodeFacade_GetHeartbeats(t *testing.T) {
	t.Parallel()

	node := &mock.NodeStub{
		GetHeartbeatsHandler: func() []data.PubKeyHeartbeat {
			return []data.PubKeyHeartbeat{
				{
					PublicKey:       "pk1",
					TimeStamp:       time.Now(),
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
				{
					PublicKey:       "pk2",
					TimeStamp:       time.Now(),
					IsActive:        true,
					ReceivedShardID: uint32(0),
				},
			}
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	result, err := nf.GetHeartbeats()

	assert.Nil(t, err)
	fmt.Println(result)
}

func TestNodeFacade_GetDataValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		ExecuteSCQueryHandler: func(query *process.SCQuery) (*vmcommon.VMOutput, error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil
		},
	}
	nf, err := NewNodeFacade(arg)
	require.NoError(t, err)

	_, _ = nf.ExecuteSCQuery(nil)
	assert.True(t, wasCalled)
}

func TestNodeFacade_EmptyRestInterface(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = ""
	nf, _ := NewNodeFacade(arg)

	assert.Equal(t, DefaultRestInterface, nf.RestApiInterface())
}

func TestNodeFacade_RestInterface(t *testing.T) {
	t.Parallel()

	intf := "localhost:1111"
	arg := createMockArguments()
	arg.FacadeConfig.RestApiInterface = intf
	nf, _ := NewNodeFacade(arg)

	assert.Equal(t, intf, nf.RestApiInterface())
}

func TestNodeFacade_ValidatorStatisticsApi(t *testing.T) {
	t.Parallel()

	mapToRet := make(map[string]*state.ValidatorApiResponse)
	mapToRet["test"] = &state.ValidatorApiResponse{NumLeaderFailure: 5}
	node := &mock.NodeStub{
		ValidatorStatisticsApiCalled: func() (map[string]*state.ValidatorApiResponse, error) {
			return mapToRet, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	res, err := nf.ValidatorStatisticsApi()
	assert.Nil(t, err)
	assert.Equal(t, mapToRet, res)
}

func TestNodeFacade_SendBulkTransactions(t *testing.T) {
	t.Parallel()

	expectedNumOfSuccessfulTxs := uint64(1)
	sendBulkTxsWasCalled := false
	node := &mock.NodeStub{
		SendBulkTransactionsHandler: func(txs []*transaction.Transaction) (uint64, error) {
			sendBulkTxsWasCalled = true
			return expectedNumOfSuccessfulTxs, nil
		},
	}

	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	txs := make([]*transaction.Transaction, 0)
	txs = append(txs, &transaction.Transaction{Nonce: 1})

	res, err := nf.SendBulkTransactions(txs)
	assert.Nil(t, err)
	assert.Equal(t, expectedNumOfSuccessfulTxs, res)
	assert.True(t, sendBulkTxsWasCalled)
}

func TestNodeFacade_StatusMetrics(t *testing.T) {
	t.Parallel()

	apiResolverMetricsRequested := false
	apiResStub := &mock.ApiResolverStub{
		StatusMetricsHandler: func() external.StatusMetricsHandler {
			apiResolverMetricsRequested = true
			return nil
		},
	}

	arg := createMockArguments()
	arg.ApiResolver = apiResStub
	nf, _ := NewNodeFacade(arg)

	_ = nf.StatusMetrics()

	assert.True(t, apiResolverMetricsRequested)
}

func TestNodeFacade_PprofEnabled(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.FacadeConfig.PprofEnabled = true
	nf, _ := NewNodeFacade(arg)

	assert.True(t, nf.PprofEnabled())
}

func TestNodeFacade_RestAPIServerDebugMode(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.RestAPIServerDebugMode = true
	nf, _ := NewNodeFacade(arg)

	assert.True(t, nf.RestAPIServerDebugMode())
}

func TestNodeFacade_CreateTransaction(t *testing.T) {
	t.Parallel()

	nodeCreateTxWasCalled := false
	node := &mock.NodeStub{
		CreateTransactionHandler: func(_ uint64, _ string, _ string, _ []byte, _ string, _ []byte, _ uint64, _ uint64, _ []byte, _ string, _ string, _, _ uint32) (*transaction.Transaction, []byte, error) {
			nodeCreateTxWasCalled = true
			return nil, nil, nil
		},
	}
	arg := createMockArguments()
	arg.Node = node
	nf, _ := NewNodeFacade(arg)

	_, _, _ = nf.CreateTransaction(0, "0", "0", nil, "0", nil, 0, 0, []byte("0"), "0", "chainID", 1, 0)

	assert.True(t, nodeCreateTxWasCalled)
}

func TestNodeFacade_Trigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	expectedErr := errors.New("expected err")
	arg := createMockArguments()
	epoch := uint32(4638)
	recoveredEpoch := uint32(0)
	recoveredWithEarlyEndOfEpoch := atomicCore.Flag{}
	arg.Node = &mock.NodeStub{
		DirectTriggerCalled: func(epoch uint32, withEarlyEndOfEpoch bool) error {
			wasCalled = true
			atomic.StoreUint32(&recoveredEpoch, epoch)
			recoveredWithEarlyEndOfEpoch.SetValue(withEarlyEndOfEpoch)

			return expectedErr
		},
	}
	nf, _ := NewNodeFacade(arg)

	err := nf.Trigger(epoch, true)

	assert.True(t, wasCalled)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, epoch, atomic.LoadUint32(&recoveredEpoch))
	assert.True(t, recoveredWithEarlyEndOfEpoch.IsSet())
}

func TestNodeFacade_IsSelfTrigger(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		IsSelfTriggerCalled: func() bool {
			wasCalled = true
			return true
		},
	}
	nf, _ := NewNodeFacade(arg)

	isSelf := nf.IsSelfTrigger()

	assert.True(t, wasCalled)
	assert.True(t, isSelf)
}

func TestNodeFacade_EncodeDecodeAddressPubkey(t *testing.T) {
	t.Parallel()

	buff := []byte("abcdefg")
	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)
	encoded, err := nf.EncodeAddressPubkey(buff)
	assert.Nil(t, err)

	recoveredBytes, err := nf.DecodeAddressPubkey(encoded)

	assert.Nil(t, err)
	assert.Equal(t, buff, recoveredBytes)
}

func TestElrondNodeFacade_GetQueryHandler(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetQueryHandlerCalled: func(name string) (handler debug.QueryHandler, err error) {
			wasCalled = true

			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	qh, err := nf.GetQueryHandler("")

	assert.Nil(t, qh)
	assert.Nil(t, err)
	assert.True(t, wasCalled)
}

func TestNodeFacade_GetPeerInfo(t *testing.T) {
	t.Parallel()

	pinfo := core.QueryP2PPeerInfo{
		Pid: "pid",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetPeerInfoCalled: func(pid string) ([]core.QueryP2PPeerInfo, error) {
			return []core.QueryP2PPeerInfo{pinfo}, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	val, err := nf.GetPeerInfo("")

	assert.Nil(t, err)
	assert.Equal(t, []core.QueryP2PPeerInfo{pinfo}, val)
}

func TestNodeFacade_GetThrottlerForEndpointNoConfigShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{} // ensure it is empty
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("any-endpoint")

	assert.Nil(t, thr)
	assert.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointNotFoundShouldReturnNilAndFalse(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("different-endpoint")

	assert.Nil(t, thr)
	assert.False(t, ok)
}

func TestNodeFacade_GetThrottlerForEndpointShouldFindAndReturn(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.WsAntifloodConfig.EndpointsThrottlers = []config.EndpointsThrottlersConfig{
		{
			Endpoint:         "endpoint",
			MaxNumGoRoutines: 10,
		},
	}
	nf, _ := NewNodeFacade(arg)

	thr, ok := nf.GetThrottlerForEndpoint("endpoint")

	assert.NotNil(t, thr)
	assert.True(t, ok)
}

func TestNodeFacade_GetKeyValuePairs(t *testing.T) {
	t.Parallel()

	expectedPairs := map[string]string{"k": "v"}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetKeyValuePairsCalled: func(address string, _ context.Context) (map[string]string, error) {
			return expectedPairs, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetKeyValuePairs("addr")
	assert.NoError(t, err)
	assert.Equal(t, expectedPairs, res)
}

func TestNodeFacade_GetAllESDTTokens(t *testing.T) {
	t.Parallel()

	expectedTokens := map[string]*esdt.ESDigitalToken{
		"token0": {Value: big.NewInt(10)},
		"token1": {TokenMetaData: &esdt.MetaData{Name: []byte("name1")}},
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllESDTTokensCalled: func(_ string, _ context.Context) (map[string]*esdt.ESDigitalToken, error) {
			return expectedTokens, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetAllESDTTokens("addr")
	assert.NoError(t, err)
	assert.Equal(t, expectedTokens, res)
}

func TestNodeFacade_GetESDTData(t *testing.T) {
	t.Parallel()

	expectedData := &esdt.ESDigitalToken{
		TokenMetaData: &esdt.MetaData{Name: []byte("name1")},
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetESDTDataCalled: func(_ string, _ string, _ uint64) (*esdt.ESDigitalToken, error) {
			return expectedData, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetESDTData("addr", "tkn", 0)
	assert.NoError(t, err)
	assert.Equal(t, expectedData, res)
}

func TestNodeFacade_GetValueForKey(t *testing.T) {
	t.Parallel()

	expectedValue := "value"
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetValueForKeyCalled: func(_ string, _ string) (string, error) {
			return expectedValue, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetValueForKey("addr", "key")
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, res)
}

func TestNodeFacade_GetAllIssuedESDTs(t *testing.T) {
	t.Parallel()

	expectedValue := []string{"value"}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllIssuedESDTsCalled: func(_ string, _ context.Context) ([]string, error) {
			return expectedValue, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	res, err := nf.GetAllIssuedESDTs("")
	assert.NoError(t, err)
	assert.Equal(t, expectedValue, res)
}

func TestNodeFacade_GetESDTsWithRole(t *testing.T) {
	t.Parallel()

	expectedResponse := []string{"ABC-1q2w3e", "PPP-sc78gh"}
	args := createMockArguments()

	args.Node = &mock.NodeStub{
		GetESDTsWithRoleCalled: func(address string, role string, _ context.Context) ([]string, error) {
			return expectedResponse, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	res, err := nf.GetESDTsWithRole("address", "role")
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)
}

func TestNodeFacade_GetNFTTokenIDsRegisteredByAddress(t *testing.T) {
	t.Parallel()

	expectedResponse := []string{"ABC-1q2w3e", "PPP-sc78gh"}
	args := createMockArguments()

	args.Node = &mock.NodeStub{
		GetNFTTokenIDsRegisteredByAddressCalled: func(address string, _ context.Context) ([]string, error) {
			return expectedResponse, nil
		},
	}

	nf, _ := NewNodeFacade(args)

	res, err := nf.GetNFTTokenIDsRegisteredByAddress("address")
	require.NoError(t, err)
	require.Equal(t, expectedResponse, res)
}

func TestNodeFacade_GetAllIssuedESDTsWithError(t *testing.T) {
	t.Parallel()

	localErr := errors.New("local")
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetAllIssuedESDTsCalled: func(_ string, _ context.Context) ([]string, error) {
			return nil, localErr
		},
	}

	nf, _ := NewNodeFacade(arg)

	_, err := nf.GetAllIssuedESDTs("")
	assert.Equal(t, err, localErr)
}

func TestNodeFacade_ValidateTransactionForSimulation(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		ValidateTransactionForSimulationCalled: func(tx *transaction.Transaction, bypassSignature bool) error {
			called = true
			return nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	err := nf.ValidateTransactionForSimulation(&transaction.Transaction{}, false)
	assert.Nil(t, err)
	assert.True(t, called)
}

func TestNodeFacade_GetTotalStakedValue(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetTotalStakedValueHandler: func(ctx context.Context) (*api.StakeValues, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetTotalStakedValue()

	assert.Nil(t, err)
	assert.True(t, called)
}

func TestNodeFacade_GetDelegatorsList(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetDelegatorsListHandler: func(ctx context.Context) ([]*api.Delegator, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetDelegatorsList()

	assert.Nil(t, err)
	assert.True(t, called)
}

func TestNodeFacade_GetDirectStakedList(t *testing.T) {
	t.Parallel()

	called := false
	arg := createMockArguments()
	arg.ApiResolver = &mock.ApiResolverStub{
		GetDirectStakedListHandler: func(ctx context.Context) ([]*api.DirectStakedValue, error) {
			called = true
			return nil, nil
		},
	}
	nf, _ := NewNodeFacade(arg)
	_, err := nf.GetDirectStakedList()

	assert.Nil(t, err)
	assert.True(t, called)
}

func TestNodeFacade_GetProofCurrentRootHashIsEmptyShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Blockchain = &testscommon.ChainHandlerStub{
		GetCurrentBlockRootHashCalled: func() []byte {
			return nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProofCurrentRootHash("addr")
	assert.Nil(t, response)
	assert.Equal(t, ErrEmptyRootHash, err)
}

func TestNodeFacade_GetProof(t *testing.T) {
	t.Parallel()

	expectedResponse := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof")},
		Value:    []byte("value"),
		RootHash: "rootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofCalled: func(_ string, _ string) (*common.GetProofResponse, error) {
			return expectedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProof("hash", "addr")
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestNodeFacade_GetProofCurrentRootHash(t *testing.T) {
	t.Parallel()

	expectedResponse := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof")},
		Value:    []byte("value"),
		RootHash: "rootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofCalled: func(_ string, _ string) (*common.GetProofResponse, error) {
			return expectedResponse, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.GetProofCurrentRootHash("addr")
	assert.Nil(t, err)
	assert.Equal(t, expectedResponse, response)
}

func TestNodeFacade_GetProofDataTrie(t *testing.T) {
	t.Parallel()

	expectedResponseMainTrie := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof"), []byte("mainTrie")},
		Value:    []byte("accountBytes"),
		RootHash: "rootHash",
	}
	expectedResponseDataTrie := &common.GetProofResponse{
		Proof:    [][]byte{[]byte("valid"), []byte("proof"), []byte("dataTrie")},
		Value:    []byte("value"),
		RootHash: "dataTrieRootHash",
	}
	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		GetProofDataTrieCalled: func(_ string, _ string, _ string) (*common.GetProofResponse, *common.GetProofResponse, error) {
			return expectedResponseMainTrie, expectedResponseDataTrie, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	mainTrieResponse, dataTrieResponse, err := nf.GetProofDataTrie("hash", "addr", "key")
	assert.Nil(t, err)
	assert.Equal(t, expectedResponseMainTrie, mainTrieResponse)
	assert.Equal(t, expectedResponseDataTrie, dataTrieResponse)
}

func TestNodeFacade_VerifyProof(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	arg.Node = &mock.NodeStub{
		VerifyProofCalled: func(_ string, _ string, _ [][]byte) (bool, error) {
			return true, nil
		},
	}
	nf, _ := NewNodeFacade(arg)

	response, err := nf.VerifyProof("hash", "addr", [][]byte{[]byte("proof")})
	assert.Nil(t, err)
	assert.True(t, response)
}

func TestNodeFacade_ExecuteSCQuery(t *testing.T) {
	t.Parallel()

	executeScQueryHandlerWasCalled := false
	arg := createMockArguments()

	expectedAddress := []byte("addr")
	expectedBalance := big.NewInt(37)
	expectedVmOutput := &vmcommon.VMOutput{
		ReturnData: [][]byte{[]byte("test return data")},
		ReturnCode: vmcommon.AccountCollision,
		OutputAccounts: map[string]*vmcommon.OutputAccount{
			"key0": {
				Address: expectedAddress,
				Balance: expectedBalance,
			},
		},
	}
	arg.ApiResolver = &mock.ApiResolverStub{
		ExecuteSCQueryHandler: func(_ *process.SCQuery) (*vmcommon.VMOutput, error) {
			executeScQueryHandlerWasCalled = true
			return expectedVmOutput, nil
		},
	}

	nf, _ := NewNodeFacade(arg)

	apiVmOutput, err := nf.ExecuteSCQuery(&process.SCQuery{})
	require.NoError(t, err)
	require.True(t, executeScQueryHandlerWasCalled)
	require.Equal(t, expectedVmOutput.ReturnData, apiVmOutput.ReturnData)
	require.Equal(t, expectedVmOutput.ReturnCode.String(), apiVmOutput.ReturnCode)
	require.Equal(t, 1, len(apiVmOutput.OutputAccounts))

	outputAccount := apiVmOutput.OutputAccounts[hex.EncodeToString([]byte("key0"))]
	require.Equal(t, expectedBalance, outputAccount.Balance)
	require.Equal(t, hex.EncodeToString(expectedAddress), outputAccount.Address)
}

func TestNodeFacade_GetBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &api.Block{
		Round: 1,
		Nonce: 2,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetBlockByRoundCalled: func(_ uint64, _ bool) (*api.Block, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetBlockByRound(0, false)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

// ---- MetaBlock

func TestNodeFacade_GetInternalMetaBlockByNonceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByNonce(common.ApiOutputFormatProto, 0)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMetaBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByRound(common.ApiOutputFormatProto, 0)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMetaBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MetaBlock{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMetaBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMetaBlockByHash(common.ApiOutputFormatProto, "dummyhash")

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

// ---- ShardBlock

func TestNodeFacade_GetInternalShardBlockByNonceShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByNonceCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByNonce(common.ApiOutputFormatProto, 0)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalShardBlockByRoundShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByRoundCalled: func(_ common.ApiOutputFormat, _ uint64) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByRound(common.ApiOutputFormatProto, 0)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalShardBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.Header{
		Round: 0,
		Nonce: 0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalShardBlockByHashCalled: func(_ common.ApiOutputFormat, _ string) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalShardBlockByHash(common.ApiOutputFormatProto, "dummyhash")

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestNodeFacade_GetInternalMiniBlockByHashShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockArguments()
	blk := &block.MiniBlock{
		ReceiverShardID: 0,
		SenderShardID:   0,
	}

	arg.ApiResolver = &mock.ApiResolverStub{
		GetInternalMiniBlockCalled: func(_ common.ApiOutputFormat, _ string, epoch uint32) (interface{}, error) {
			return blk, nil
		},
	}

	nf, _ := NewNodeFacade(arg)
	ret, err := nf.GetInternalMiniBlockByHash(common.ApiOutputFormatProto, "dummyhash", 1)

	assert.Nil(t, err)
	assert.Equal(t, ret, blk)
}

func TestFacade_convertVmOutputToApiResponse(t *testing.T) {
	arg := createMockArguments()
	nf, _ := NewNodeFacade(arg)

	convertAddressFunc := func(input []byte) string {
		// NodeStub uses hex.EncodeToString for the stubbed EncodeAddressPubkey() function
		return hex.EncodeToString(input)
	}

	retData := [][]byte{[]byte("ret_data_0")}
	outAcc, outAccStorageKey, outAccOffset := []byte("addr0"), []byte("out_acc_storage_key"), []byte("offset")
	outAccTransferSndrAddr := []byte("addr1")
	logId, logAddr, logTopics, logData := []byte("log_id"), []byte("log_addr"), [][]byte{[]byte("log_topic")}, []byte("log_data")
	vmInput := vmcommon.VMOutput{
		ReturnData: retData,
		OutputAccounts: map[string]*vmcommon.OutputAccount{
			string(outAcc): {
				Address: outAcc,
				StorageUpdates: map[string]*vmcommon.StorageUpdate{
					string(outAccStorageKey): {
						Offset: outAccOffset,
					},
				},
				OutputTransfers: []vmcommon.OutputTransfer{
					{
						SenderAddress: outAccTransferSndrAddr,
					},
				},
			},
		},
		Logs: []*vmcommon.LogEntry{
			{
				Identifier: logId,
				Address:    logAddr,
				Topics:     logTopics,
				Data:       logData,
			},
		},
	}

	expectedOutputAccounts := map[string]*vm.OutputAccountApi{
		convertAddressFunc(outAcc): {
			Address: convertAddressFunc(outAcc),
			StorageUpdates: map[string]*vm.StorageUpdateApi{
				hex.EncodeToString(outAccStorageKey): {
					Offset: outAccOffset,
				},
			},
			OutputTransfers: []vm.OutputTransferApi{
				{
					SenderAddress: convertAddressFunc(outAccTransferSndrAddr),
				},
			},
		},
	}

	expectedLogs := []*vm.LogEntryApi{
		{
			Identifier: logId,
			Address:    convertAddressFunc(logAddr),
			Topics:     logTopics,
			Data:       logData,
		},
	}

	res := nf.convertVmOutputToApiResponse(&vmInput)
	require.Equal(t, retData, res.ReturnData)
	require.Equal(t, expectedOutputAccounts, res.OutputAccounts)
	require.Equal(t, expectedLogs, res.Logs)
}

func TestNodeFacade_GetTransactionsPool(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedErr := errors.New("expected error")
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolCalled: func() (*common.TransactionsPoolAPIResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPool()
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedPool := &common.TransactionsPoolAPIResponse{
			RegularTransactions:  []string{"tx0", "tx1"},
			SmartContractResults: []string{"tx2", "tx3"},
			Rewards:              []string{"tx4"},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolCalled: func() (*common.TransactionsPoolAPIResponse, error) {
				return expectedPool, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPool()
		require.NoError(t, err)
		require.Equal(t, expectedPool, res)
	})
}

func TestNodeFacade_GetTransactionsPoolForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedErr := errors.New("expected error")
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolForSenderCalled: func(sender string) (*common.TransactionsPoolForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolForSender("")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedSender := "alice"
		expectedResponse := &common.TransactionsPoolForSenderApiResponse{
			Sender:       expectedSender,
			Transactions: []string{"txhash1", "txhash2"},
		}
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolForSenderCalled: func(sender string) (*common.TransactionsPoolForSenderApiResponse, error) {
				require.Equal(t, expectedSender, sender)
				return expectedResponse, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedResponse, res)
	})
}

func TestNodeFacade_GetLastPoolNonceForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedErr := errors.New("expected error")
		arg.ApiResolver = &mock.ApiResolverStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return 0, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetLastPoolNonceForSender("")
		require.Equal(t, uint64(0), res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedSender := "alice"
		expectedNonce := uint64(33)
		arg.ApiResolver = &mock.ApiResolverStub{
			GetLastPoolNonceForSenderCalled: func(sender string) (uint64, error) {
				return expectedNonce, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetLastPoolNonceForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedNonce, res)
	})
}

func TestNodeFacade_GetTransactionsPoolNonceGapsForSender(t *testing.T) {
	t.Parallel()

	t.Run("should error", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
		expectedErr := errors.New("expected error")
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return nil, expectedErr
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolNonceGapsForSender("")
		require.Nil(t, res)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		arg := createMockArguments()
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
		arg.ApiResolver = &mock.ApiResolverStub{
			GetTransactionsPoolNonceGapsForSenderCalled: func(sender string) (*common.TransactionsPoolNonceGapsForSenderApiResponse, error) {
				return expectedNonceGaps, nil
			},
		}

		nf, _ := NewNodeFacade(arg)
		res, err := nf.GetTransactionsPoolNonceGapsForSender(expectedSender)
		require.NoError(t, err)
		require.Equal(t, expectedNonceGaps, res)
	})
}
