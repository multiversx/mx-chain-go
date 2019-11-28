package external_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeApiResolver_NilScDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(nil, &mock.StatusMetricsStub{})

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilScDataGetter, err)
}

func TestNewNodeApiResolver_NilStatusMetricsShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.ScDataGetterStub{}, nil)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilStatusMetrics, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.ScDataGetterStub{}, &mock.StatusMetricsStub{})

	assert.NotNil(t, nar)
	assert.Nil(t, err)
}

func TestNodeApiResolver_GetDataValueShouldCall(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(&mock.ScDataGetterStub{
		GetCalled: func(scAddress []byte, funcName string, args ...[]byte) (bytes []byte, e error) {
			wasCalled = true
			return make([]byte, 0), nil
		},
	},
		&mock.StatusMetricsStub{})

	_, _ = nar.GetVmValue("", "")

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.ScDataGetterStub{},
		&mock.StatusMetricsStub{
			StatusMetricsMapCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		})
	_ = nar.StatusMetrics().StatusMetricsMap()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusMetricsMapWhitoutP2PShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.ScDataGetterStub{},
		&mock.StatusMetricsStub{
			StatusMetricsMapWithoutP2PCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		})
	_ = nar.StatusMetrics().StatusMetricsMapWithoutP2P()

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_StatusP2PMetricsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.ScDataGetterStub{},
		&mock.StatusMetricsStub{
			StatusP2pMetricsMapCalled: func() map[string]interface{} {
				wasCalled = true
				return nil
			},
		})
	_ = nar.StatusMetrics().StatusP2pMetricsMap()

	assert.True(t, wasCalled)
}
