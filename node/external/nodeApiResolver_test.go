package external_test

import (
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/vmcommon"
	"github.com/ElrondNetwork/elrond-go/data/api"
	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/stretchr/testify/assert"
)

func createMockAgrs() external.ArgNodeApiResolver {
	return external.ArgNodeApiResolver{
		SCQueryService:          &mock.SCQueryServiceStub{},
		StatusMetricsHandler:    &mock.StatusMetricsStub{},
		TxCostHandler:           &mock.TransactionCostEstimatorMock{},
		TotalStakedValueHandler: &mock.StakeValuesProcessorStub{},
		DirectStakedListHandler: &mock.DirectStakedListProcessorStub{},
		DelegatedListHandler:    &mock.DelegatedListProcessorStub{},
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

func TestNewNodeApiResolver_NilTransactionCostEstimator(t *testing.T) {
	t.Parallel()

	arg := createMockAgrs()
	arg.TxCostHandler = nil
	nar, err := external.NewNodeApiResolver(arg)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilTransactionCostHandler, err)
}

func TestNewNodeApiResolver_NilVmFactory(t *testing.T) {
	t.Parallel()

	totalStakedAPIHandler, _ := disabled.NewDisabledStakeValuesProcessor()
	args := external.ApiResolverArgs{
		ScQueryService:     &mock.SCQueryServiceStub{},
		StatusMetrics:      &mock.StatusMetricsStub{},
		TxCostHandler:      &mock.TransactionCostEstimatorMock{},
		VmFactory:          nil,
		VmContainer:        &mock.VMContainerMock{},
		StakedValueHandler: totalStakedAPIHandler,
	}
	nar, err := external.NewNodeApiResolver(args)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilVmFactory, err)
}

func TestNewNodeApiResolver_NilVmContainer(t *testing.T) {
	t.Parallel()

	totalStakedAPIHandler, _ := disabled.NewDisabledStakeValuesProcessor()
	args := external.ApiResolverArgs{
		ScQueryService:     &mock.SCQueryServiceStub{},
		StatusMetrics:      &mock.StatusMetricsStub{},
		TxCostHandler:      &mock.TransactionCostEstimatorMock{},
		VmFactory:          &mock.VmMachinesContainerFactoryMock{},
		VmContainer:        nil,
		StakedValueHandler: totalStakedAPIHandler,
	}
	nar, err := external.NewNodeApiResolver(args)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilVmContainer, err)
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

func TestNodeApiResolver_Close_ShouldWork(t *testing.T) {
	t.Parallel()

	calledClose := false
	totalStakedAPIHandler, _ := disabled.NewDisabledStakeValuesProcessor()
	args := external.ApiResolverArgs{
		ScQueryService: &mock.SCQueryServiceStub{},
		StatusMetrics:  &mock.StatusMetricsStub{},
		TxCostHandler:  &mock.TransactionCostEstimatorMock{},
		VmFactory:      &mock.VmMachinesContainerFactoryMock{},
		VmContainer: &mock.VMContainerMock{
			CloseCalled: func() error {
				calledClose = true
				return nil
			},
		},
		StakedValueHandler: totalStakedAPIHandler,
	}

	nar, _ := external.NewNodeApiResolver(args)

	err := nar.Close()
	assert.Nil(t, err)
	assert.True(t, calledClose)
}

func TestNodeApiResolver_Close_OnErrorShouldError(t *testing.T) {
	t.Parallel()

	totalStakedAPIHandler, _ := disabled.NewDisabledStakeValuesProcessor()
	args := external.ApiResolverArgs{
		ScQueryService: &mock.SCQueryServiceStub{},
		StatusMetrics:  &mock.StatusMetricsStub{},
		TxCostHandler:  &mock.TransactionCostEstimatorMock{},
		VmFactory:      &mock.VmMachinesContainerFactoryMock{},
		VmContainer: &mock.VMContainerMock{
			CloseCalled: func() error {
				return fmt.Errorf("error")
			},
		},
		StakedValueHandler: totalStakedAPIHandler,
	}
	nar, _ := external.NewNodeApiResolver(args)

	err := nar.Close()
	assert.NotNil(t, err)
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

func TestNodeApiResolver_GetTotalStakedValue(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockAgrs()
	stakeValue := &api.StakeValues{}
	arg.TotalStakedValueHandler = &mock.StakeValuesProcessorStub{
		GetTotalStakedValueCalled: func() (*api.StakeValues, error) {
			wasCalled = true
			return stakeValue, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredStakeValue, err := nar.GetTotalStakedValue()
	assert.Nil(t, err)
	assert.True(t, recoveredStakeValue == stakeValue) //pointer testing
	assert.True(t, wasCalled)
}

func TestNodeApiResolver_GetDelegatorsList(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockAgrs()
	delegators := make([]*api.Delegator, 1)
	arg.DelegatedListHandler = &mock.DelegatedListProcessorStub{
		GetDelegatorsListCalled: func() ([]*api.Delegator, error) {
			wasCalled = true
			return delegators, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredDelegatorsList, err := nar.GetDelegatorsList()
	assert.Nil(t, err)
	assert.Equal(t, recoveredDelegatorsList, delegators)
	assert.True(t, wasCalled)
}

func TestNodeApiResolver_GetDirectStakedList(t *testing.T) {
	t.Parallel()

	wasCalled := false
	arg := createMockAgrs()
	directStakedValueList := make([]*api.DirectStakedValue, 1)
	arg.DirectStakedListHandler = &mock.DirectStakedListProcessorStub{
		GetDirectStakedListCalled: func() ([]*api.DirectStakedValue, error) {
			wasCalled = true
			return directStakedValueList, nil
		},
	}

	nar, _ := external.NewNodeApiResolver(arg)
	recoveredDirectStakedValueList, err := nar.GetDirectStakedList()
	assert.Nil(t, err)
	assert.Equal(t, recoveredDirectStakedValueList, directStakedValueList)
	assert.True(t, wasCalled)
}
