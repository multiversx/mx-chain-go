package dataRetriever_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/dataRetriever/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestDataType_StringHashType(t *testing.T) {
	t.Parallel()

	requestDataType := dataRetriever.HashType
	rd := requestDataType.String()

	assert.Equal(t, "hash type", rd)
}

func TestRequestDataType_StringHashArrayType(t *testing.T) {
	t.Parallel()

	requestDataType := dataRetriever.HashArrayType
	rd := requestDataType.String()

	assert.Equal(t, "hash array type", rd)
}

func TestRequestDataType_StringNonceType(t *testing.T) {
	t.Parallel()

	requestDataType := dataRetriever.NonceType
	rd := requestDataType.String()

	assert.Equal(t, "nonce type", rd)
}

func TestRequestDataType_StringEpochType(t *testing.T) {
	t.Parallel()

	requestDataType := dataRetriever.EpochType
	rd := requestDataType.String()

	assert.Equal(t, "epoch type", rd)
}

func TestRequestDataType_UnknownType(t *testing.T) {
	t.Parallel()

	var requestData dataRetriever.RequestDataType = 6
	rd := requestData.String()

	assert.Equal(t, fmt.Sprintf("unknown type %d", 6), rd)
}

func TestRequestData_UnmarshalNilMarshalizer(t *testing.T) {
	t.Parallel()

	requestData := dataRetriever.RequestData{}

	err := requestData.UnmarshalWith(nil, &mock.P2PMessageMock{})
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

	err := requestData.UnmarshalWith(&mock.MarshalizerMock{}, &mock.P2PMessageMock{})
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
	}, &mock.P2PMessageMock{
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
	}, &mock.P2PMessageMock{
		DataField: []byte("data"),
	})
	require.Nil(t, err)
}
