package external_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/node/external"
	"github.com/ElrondNetwork/elrond-go/node/mock"
	"github.com/stretchr/testify/assert"
)

func TestNewNodeApiResolver_NilScDataGetterShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(nil, &mock.NodeDetailsStub{})

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilScDataGetter, err)
}

func TestNewNodeApiResolver_NilNodeDetailsShouldErr(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.ScDataGetterStub{}, nil)

	assert.Nil(t, nar)
	assert.Equal(t, external.ErrNilNodeDetails, err)
}

func TestNewNodeApiResolver_ShouldWork(t *testing.T) {
	t.Parallel()

	nar, err := external.NewNodeApiResolver(&mock.ScDataGetterStub{}, &mock.NodeDetailsStub{})

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
		&mock.NodeDetailsStub{})

	_, _ = nar.GetVmValue("", "")

	assert.True(t, wasCalled)
}

func TestNodeApiResolver_NodeDetailsMapShouldBeCalled(t *testing.T) {
	t.Parallel()

	wasCalled := false
	nar, _ := external.NewNodeApiResolver(
		&mock.ScDataGetterStub{},
		&mock.NodeDetailsStub{
			DetailsMapCalled: func() (map[string]interface{}, error) {
				wasCalled = true
				return nil, nil
			},
		})
	_, _ = nar.NodeDetails().DetailsMap()

	assert.True(t, wasCalled)
}
