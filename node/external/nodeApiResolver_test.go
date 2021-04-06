package external_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/node/trieIterators/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func createMockAgrs() external.ArgNodeApiResolver {
	return external.ArgNodeApiResolver{
		SCQueryService:          &mock.SCQueryServiceStub{},
		StatusMetricsHandler:    &mock.StatusMetricsStub{},
		TxCostHandler:           &mock.TransactionCostEstimatorMock{},
		TotalStakedValueHandler: disabled.NewDisabledStakeValuesProcessor(),
		DirectStakedListHandler: disabled.NewDisabledDirectStakedListProcessor(),
		DelegatedListHandler:    disabled.NewDisabledDelegatedListProcessor(),
	}
}

func TestNewNodeApiResolver_NilSCQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.SCQueryService = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilSCQueryService, err)
}

func TestNewNodeApiResolver_NilStatusMetricsShouldErr(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.StatusMetricsHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilStatusMetrics, err)
}

func TestNewNodeApiResolver_NilTransactionCostEstsimator(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.TxCostHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTransactionCostHandler, err)
}

func TestNewNodeApiResolver_NilTotalStakedValueHandler(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.TotalStakedValueHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTotalStakedValueHandler, err)
}

func TestNewNodeApiResolver_NilDirectStakedListHandler(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.DirectStakedListHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilDirectStakeListHandler, err)
}

func TestNewNodeApiResolver_NilDelegatedListHandler(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.DelegatedListHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilDelegatedListHandler, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, err)
	assert.False(t, check.IfNil(nar))
}

func TestNodeApiResolver_GetDataValueShouldCall(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
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

	arg := createMockAgrs()
	wasCalled := false
	arg.StatusMetricsHandler = &mock.StatusMetricsStub{
		StatusMetricsMapWithoutP2PCalled: func() map[string]interface{} {
			wasCalled = true
			return nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_ = nar.StatusMetrics().StatusMetricsMapWithoutP2P()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusP2PMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	wasCalled := false
	arg.StatusMetricsHandler = &mock.StatusMetricsStub{
		StatusP2pMetricsMapCalled: func() map[string]interface{} {
			wasCalled = true
			return nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_ = nar.StatusMetrics().StatusP2pMetricsMap()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_NetworkMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	wasCalled := false
	arg.StatusMetricsHandler = &mock.StatusMetricsStub{
		NetworkMetricsCalled: func() map[string]interface{} {
			wasCalled = true
			return nil
		},
	}
	nar, _ := external.NewNodeApiResolver(arg)
	_ = nar.StatusMetrics().NetworkMetrics()

	assert.True(t, wasCalled)
}

//TODO add more unit tests
