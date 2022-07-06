package consensus_test

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsensusMessage_NewConsensusMessageShouldWork(t *testing.T) {
	t.Parallel()

	cnsMsg := consensus.NewConsensusMessage(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		-1,
		0,
		[]byte("chain ID"),
		nil,
		nil,
		nil,
		"pid",
	)

	assert.NotNil(t, cnsMsg)
}

func TestMessageSerializing(t *testing.T) {
	t.Parallel()

	message := &consensus.Message{
		BlockHeaderHash:    []byte("blockHash1"),
		SignatureShare:     []byte("sigShare"),
		Body:               []byte{},
		Header:             []byte{},
		PubKey:             []byte{},
		Signature:          []byte{},
		MsgType:            2,
		RoundIndex:         6,
		ChainID:            []byte{},
		PubKeysBitmap:      []byte{},
		AggregateSignature: []byte{},
		LeaderSignature:    []byte{},
		OriginatorPid:      []byte("123"),
		InvalidSigners:     []byte("incadsada"),
	}

	// messageOld := &consensus.MessageOld{
	// 	BlockHeaderHash:    []byte{},
	// 	SignatureShare:     []byte{},
	// 	Body:               []byte{},
	// 	Header:             []byte{},
	// 	PubKey:             []byte{},
	// 	Signature:          []byte{},
	// 	MsgType:            0,
	// 	RoundIndex:         0,
	// 	ChainID:            []byte{},
	// 	PubKeysBitmap:      []byte{},
	// 	AggregateSignature: []byte{},
	// 	LeaderSignature:    []byte{},
	// 	OriginatorPid:      []byte{},
	// }

	marshaller := &testscommon.MarshalizerMock{}

	msgOldBytes, _ := marshaller.Marshal(message)

	newM := &consensus.MessageOld{}
	err := marshaller.Unmarshal(newM, msgOldBytes)
	require.NotNil(t, newM.InvalidSigners)
	require.Nil(t, err)

	newM2 := &consensus.Message{}
	err = marshaller.Unmarshal(newM2, msgOldBytes)
	require.NotNil(t, newM2.InvalidSigners)
	require.Nil(t, err)
}
