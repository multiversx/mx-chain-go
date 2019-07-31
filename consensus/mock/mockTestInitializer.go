package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
)

func InitChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &ChronologyHandlerMock{}
	return chr
}

func InitBlockProcessorMock() *BlockProcessorMock {
	blockProcessorMock := &BlockProcessorMock{}
	blockProcessorMock.CreateBlockCalled = func(round uint64, haveTime func() bool) (data.BodyHandler, error) {
		emptyBlock := make(block.Body, 0)

		return emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertAccountStateCalled = func() {}
	blockProcessorMock.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return nil
	}
	blockProcessorMock.CreateBlockHeaderCalled = func(body data.BodyHandler, round uint64, haveTime func() bool) (header data.HeaderHandler, e error) {
		return &block.Header{RootHash: []byte{}}, nil
	}
	blockProcessorMock.DecodeBlockBodyCalled = func(dta []byte) data.BodyHandler {
		return block.Body{}
	}
	blockProcessorMock.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
		return &block.Header{}
	}
	blockProcessorMock.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte, 0), make(map[string][][]byte, 0), nil
	}

	return blockProcessorMock
}

func InitMultiSignerMock() *BelNevMock {
	multiSigner := NewMultiSigner()
	multiSigner.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("commitment")
	}
	multiSigner.VerifySignatureShareMock = func(index uint16, sig []byte, msg []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.VerifyMock = func(msg []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}
	multiSigner.AggregateCommitmentsMock = func(bitmap []byte) error {
		return nil
	}
	multiSigner.CreateSignatureShareMock = func(msg []byte, bitmap []byte) ([]byte, error) {
		return []byte("partialSign"), nil
	}
	return multiSigner
}

func InitKeys() (*KeyGenMock, *PrivateKeyMock, *PublicKeyMock) {
	toByteArrayMock := func() ([]byte, error) {
		return []byte("byteArray"), nil
	}
	privKeyMock := &PrivateKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	pubKeyMock := &PublicKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	privKeyFromByteArr := func(b []byte) (crypto.PrivateKey, error) {
		return privKeyMock, nil
	}
	pubKeyFromByteArr := func(b []byte) (crypto.PublicKey, error) {
		return pubKeyMock, nil
	}
	keyGenMock := &KeyGenMock{
		PrivateKeyFromByteArrayMock: privKeyFromByteArr,
		PublicKeyFromByteArrayMock:  pubKeyFromByteArr,
	}
	return keyGenMock, privKeyMock, pubKeyMock
}

func InitConsensusCore() *ConsensusCoreMock {

	blockChain := &BlockChainMock{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	blockProcessorMock := InitBlockProcessorMock()
	blockTrackerMock := &BlocksTrackerMock{
		SetBlockBroadcastRoundCalled: func(nonce uint64, round int64) {
		},
	}
	bootstrapperMock := &BootstrapperMock{}
	broadcastMessengerMock := &BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return nil
		},
	}

	chronologyHandlerMock := InitChronologyHandlerMock()
	hasherMock := HasherMock{}
	marshalizerMock := MarshalizerMock{}
	blsPrivateKeyMock := &PrivateKeyMock{}
	blsSingleSignerMock := &SingleSignerMock{
		SignStub: func(private crypto.PrivateKey, msg []byte) (bytes []byte, e error) {
			return make([]byte, 0), nil
		},
	}
	multiSignerMock := InitMultiSignerMock()
	rounderMock := &RounderMock{}
	shardCoordinatorMock := ShardCoordinatorMock{}
	syncTimerMock := &SyncTimerMock{}
	validatorGroupSelector := ValidatorGroupSelectorMock{}

	container := &ConsensusCoreMock{
		blockChain,
		blockProcessorMock,
		blockTrackerMock,
		bootstrapperMock,
		broadcastMessengerMock,
		chronologyHandlerMock,
		hasherMock,
		marshalizerMock,
		blsPrivateKeyMock,
		blsSingleSignerMock,
		multiSignerMock,
		rounderMock,
		shardCoordinatorMock,
		syncTimerMock,
		validatorGroupSelector,
	}

	return container
}
