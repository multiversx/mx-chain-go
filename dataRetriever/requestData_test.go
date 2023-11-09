package dataRetriever_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-chain-go/dataRetriever"
	"github.com/multiversx/mx-chain-go/dataRetriever/mock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestDataType_StringVals(t *testing.T) {
	t.Parallel()

	tcs := []struct {
		r dataRetriever.RequestDataType
		s string
	}{
		{dataRetriever.HashType, "HashType"},
		{dataRetriever.HashArrayType, "HashArrayType"},
		{dataRetriever.NonceType, "NonceType"},
		{dataRetriever.EpochType, "EpochType"},
	}

	for _, tc := range tcs {
		t.Run(tc.s, func(t *testing.T) {
			rd := tc.r.String()
			assert.Equal(t, tc.s, rd)
		})
	}
}

func TestRequestDataType_UnknownType(t *testing.T) {
	t.Parallel()

	var requestData dataRetriever.RequestDataType = 6
	rd := requestData.String()

	assert.Equal(t, fmt.Sprintf("%d", 6), rd)
}

func TestRequestData_UnmarshalNilMarshalizer(t *testing.T) {
	t.Parallel()

	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(nil, &p2pmocks.P2PMessageMock{})
	require.Equal(t, dataRetriever.ErrNilMarshalizer, err)
}

func TestRequestData_UnmarshalNilMessageP2P(t *testing.T) {
	t.Parallel()

	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(&mock.MarshalizerMock{}, nil)
	require.Equal(t, dataRetriever.ErrNilMessage, err)
}

func TestRequestData_UnmarshalNilMessageData(t *testing.T) {
	t.Parallel()

	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(&mock.MarshalizerMock{}, &p2pmocks.P2PMessageMock{})
	require.Equal(t, dataRetriever.ErrNilDataToProcess, err)
}

func TestRequestData_CannotUnmarshal(t *testing.T) {
	t.Parallel()

	localErr := errors.New("err")
	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(&mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return localErr
		},
	}, &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
	})
	require.Equal(t, localErr, err)
}

func TestRequestData_UnmarshalOk(t *testing.T) {
	t.Parallel()

	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(&mock.MarshalizerStub{
		UnmarshalCalled: func(obj interface{}, buff []byte) error {
			return nil
		},
	}, &p2pmocks.P2PMessageMock{
		DataField: []byte("data"),
	})
	require.Nil(t, err)
}
