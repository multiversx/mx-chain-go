package external_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/ElrondNetwork/elrond-go/process"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeApiResolver_NilSCQueryServiceShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(nil, &mock.StatusMetricsStub{})

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilSCQueryService, err)
}

func TestNewNodeApiResolver_NilStatusMetricsShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.SCQueryServiceStub{}, nil)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilStatusMetrics, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.SCQueryServiceStub{}, &mock.StatusMetricsStub{})

	assert.NotNil(t, nar)
	assert.Nil(t, err)
}

func TestNodeApiResolver_GetDataValueShouldCall(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(&mock.SCQueryServiceStub{
		ExecuteQueryCalled: func(query *process.SCQuery) (vmOutput *vmcommon.VMOutput, e error) {
			wasCalled = true
			return &vmcommon.VMOutput{}, nil
		},
	},
		&mock.StatusMetricsStub{})

	_, _ = nar.ExecuteSCQuery(&process.SCQuery{
		ScAddress: []byte{0},
		FuncName:  "",
	})

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.SCQueryServiceStub{},
		&mock.StatusMetricsStub{
			StatusMetricsMapCalled: func() (map[string]interface{}, error) {
				wasCalled = true
				return nil, nil
			},
		})
	_, _ = nar.StatusMetrics().StatusMetricsMap()

	assert.True(t, wasCalled)
}
