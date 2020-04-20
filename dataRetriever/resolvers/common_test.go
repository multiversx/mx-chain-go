package resolvers_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/resolvers"
	"github.com/ElrondNetwork/elrond-go/p2p"
	"github.com/stretchr/testify/assert"
)

func createRequestMsg(dataType dataRetriever.RequestDataType, val []byte) p2p.MessageP2P {
	marshalizer := &mock.MarshalizerMock{}
	buff, _ := marshalizer.Marshal(&dataRetriever.RequestData{Type: dataType, Value: val})
	return &mock.P2PMessageMock{DataField: buff}
}

func TestProcessDebugMissingData_DisabledShouldNotCall(t *testing.T) {
	t.Parallel()

	handler := &mock.ResolverDebugHandler{
		FailedToResolveDataCalled: func(topic string, hash []byte, err error) {
			assert.Fail(t, "should have not called FailedToResolveDataCalled")
		},
		EnabledCalled: func() bool {
			return false
		},
	}

	resolvers.ProcessDebugMissingData(handler, "", make([]byte, 0), nil)
}

func TestProcessDebugMissingData_EnabledShouldCall(t *testing.T) {
	t.Parallel()

	wasCalled := false
	handler := &mock.ResolverDebugHandler{
		FailedToResolveDataCalled: func(topic string, hash []byte, err error) {
			wasCalled = true
		},
		EnabledCalled: func() bool {
			return true
		},
	}

	resolvers.ProcessDebugMissingData(handler, "", make([]byte, 0), nil)

	assert.True(t, wasCalled)
}
