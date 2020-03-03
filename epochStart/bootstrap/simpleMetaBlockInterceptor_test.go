package bootstrap_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap"
	mock2 "github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/mock"
	"github.com/ElrondNetwork/elrond-go/epochStart/mock"
	"github.com/stretchr/testify/require"
)

func TestNewSimpleMetaBlockInterceptor_NilMarshalizerShouldErr(t *testing.T) {
	t.Parallel()

	smbi, err := bootstrap.NewSimpleMetaBlockInterceptor(nil, &mock.HasherMock{})
	require.Nil(t, smbi)
	require.Equal(t, bootstrap.ErrNilMarshalizer, err)
}

func TestNewSimpleMetaBlockInterceptor_NilHasherShouldErr(t *testing.T) {
	t.Parallel()

	smbi, err := bootstrap.NewSimpleMetaBlockInterceptor(&mock.MarshalizerMock{}, nil)
	require.Nil(t, smbi)
	require.Equal(t, bootstrap.ErrNilHasher, err)
}

func TestNewSimpleMetaBlockInterceptor_OkValsShouldWork(t *testing.T) {
	t.Parallel()

	smbi, err := bootstrap.NewSimpleMetaBlockInterceptor(&mock.MarshalizerMock{}, &mock.HasherMock{})
	require.Nil(t, err)
	require.False(t, check.IfNil(smbi))
}

func TestSimpleMetaBlockInterceptor_ProcessReceivedMessage_ReceivedMessageIsNotAMetaBlockShouldNotAdd(t *testing.T) {
	t.Parallel()

	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(&mock.MarshalizerMock{}, &mock.HasherMock{})

	message := mock2.P2PMessageMock{
		DataField: []byte("not a metablock"),
	}

	_ = smbi.ProcessReceivedMessage(&message, nil)

	require.Zero(t, len(smbi.GetReceivedMetablocks()))
}

func TestSimpleMetaBlockInterceptor_ProcessReceivedMessage_UnmarshalFailsShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{Fail: true}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	mb := &block.MetaBlock{Epoch: 5}
	mbBytes, _ := marshalizer.Marshal(mb)
	message := mock2.P2PMessageMock{
		DataField: mbBytes,
	}

	_ = smbi.ProcessReceivedMessage(&message, nil)

	require.Zero(t, len(smbi.GetReceivedMetablocks()))
}

func TestSimpleMetaBlockInterceptor_ProcessReceivedMessage_ReceivedMessageIsAMetaBlockShouldAdd(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	mb := &block.MetaBlock{Epoch: 5}
	mbBytes, _ := marshalizer.Marshal(mb)
	message := mock2.P2PMessageMock{
		DataField: mbBytes,
	}

	_ = smbi.ProcessReceivedMessage(&message, nil)

	require.Equal(t, 1, len(smbi.GetReceivedMetablocks()))
}

func TestSimpleMetaBlockInterceptor_ProcessReceivedMessage_ShouldAddForMorePeers(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	mb := &block.MetaBlock{Epoch: 5}
	mbBytes, _ := marshalizer.Marshal(mb)
	message1 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer1",
	}
	message2 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer2",
	}

	_ = smbi.ProcessReceivedMessage(message1, nil)
	_ = smbi.ProcessReceivedMessage(message2, nil)

	for _, res := range smbi.GetPeersSliceForMetablocks() {
		require.Equal(t, 2, len(res))
	}
}

func TestSimpleMetaBlockInterceptor_ProcessReceivedMessage_ShouldNotAddTwiceForTheSamePeer(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	mb := &block.MetaBlock{Epoch: 5}
	mbBytes, _ := marshalizer.Marshal(mb)
	message1 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer1",
	}
	message2 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer1",
	}

	_ = smbi.ProcessReceivedMessage(message1, nil)
	_ = smbi.ProcessReceivedMessage(message2, nil)

	for _, res := range smbi.GetPeersSliceForMetablocks() {
		require.Equal(t, 1, len(res))
	}
}

func TestSimpleMetaBlockInterceptor_GetMetaBlock_NumTriesExceededShouldErr(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	// no message received, so should exit with err
	mb, err := smbi.GetMetaBlock(2, 5)
	require.Zero(t, mb)
	require.Equal(t, bootstrap.ErrNumTriesExceeded, err)
}

func TestSimpleMetaBlockInterceptor_GetMetaBlockShouldWork(t *testing.T) {
	t.Parallel()

	marshalizer := &mock.MarshalizerMock{}
	smbi, _ := bootstrap.NewSimpleMetaBlockInterceptor(marshalizer, &mock.HasherMock{})

	mb := &block.MetaBlock{Epoch: 5}
	mbBytes, _ := marshalizer.Marshal(mb)
	message1 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer1",
	}
	message2 := &mock2.P2PMessageMock{
		DataField: mbBytes,
		PeerField: "peer2",
	}

	_ = smbi.ProcessReceivedMessage(message1, nil)
	_ = smbi.ProcessReceivedMessage(message2, nil)

	mb, err := smbi.GetMetaBlock(2, 5)
	require.Nil(t, err)
	require.NotNil(t, mb)
}
