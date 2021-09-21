package mock

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
)

// InitChronologyHandlerMock -
func InitChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &ChronologyHandlerMock{}
	return chr
}

// InitBlockProcessorMock -
func InitBlockProcessorMock() *BlockProcessorMock {
	blockProcessorMock := &BlockProcessorMock{}
	blockProcessorMock.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		emptyBlock := &block.Body{}
		header.SetRootHash([]byte{})
		return header, emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertAccountStateCalled = func(header data.HeaderHandler) {}
	blockProcessorMock.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return nil
	}
	blockProcessorMock.DecodeBlockBodyCalled = func(dta []byte) data.BodyHandler {
		return &block.Body{}
	}
	blockProcessorMock.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
		return &block.Header{}
	}
	blockProcessorMock.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	blockProcessorMock.CreateNewHeaderCalled = func(round uint64, nonce uint64) data.HeaderHandler {
		return &block.Header{
			Round: round,
			Nonce: nonce,
		}
	}

	return blockProcessorMock
}

// InitMultiSignerMock -
func InitMultiSignerMock() *cryptoMocks.MultisignerMock {
	multiSigner := cryptoMocks.NewMultiSigner(21)
	multiSigner.VerifySignatureShareCalled = func(index uint16, sig []byte, msg []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.VerifyCalled = func(msg []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.AggregateSigsCalled = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}
	multiSigner.CreateSignatureShareCalled = func(msg []byte, bitmap []byte) ([]byte, error) {
		return []byte("partialSign"), nil
	}
	return multiSigner
}

// InitKeys -
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

// InitConsensusCore -
func InitConsensusCore() *ConsensusCoreMock {
	multiSignerMock := InitMultiSignerMock()

	return InitConsensusCoreWithMultiSigner(multiSignerMock)
}

// InitConsensusCoreWithMultiSigner -
func InitConsensusCoreWithMultiSigner(multiSigner crypto.MultiSigner) *ConsensusCoreMock {

	blockChain := &BlockChainMock{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	blockProcessorMock := InitBlockProcessorMock()
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
	roundHandlerMock := &RoundHandlerMock{}
	shardCoordinatorMock := ShardCoordinatorMock{}
	syncTimerMock := &SyncTimerMock{}
	validatorGroupSelector := &NodesCoordinatorMock{}
	epochStartSubscriber := &EpochStartNotifierStub{}
	antifloodHandler := &P2PAntifloodHandlerStub{}
	headerPoolSubscriber := &HeadersCacherStub{}
	peerHonestyHandler := &testscommon.PeerHonestyHandlerStub{}
	headerSigVerifier := &HeaderSigVerifierStub{}
	fallbackHeaderValidator := &testscommon.FallBackHeaderValidatorStub{}
	nodeRedundancyHandler := &NodeRedundancyHandlerStub{}

	container := &ConsensusCoreMock{
		blockChain:              blockChain,
		blockProcessor:          blockProcessorMock,
		headersSubscriber:       headerPoolSubscriber,
		bootstrapper:            bootstrapperMock,
		broadcastMessenger:      broadcastMessengerMock,
		chronologyHandler:       chronologyHandlerMock,
		hasher:                  hasherMock,
		marshalizer:             marshalizerMock,
		blsPrivateKey:           blsPrivateKeyMock,
		blsSingleSigner:         blsSingleSignerMock,
		multiSigner:             multiSigner,
		roundHandler:            roundHandlerMock,
		shardCoordinator:        shardCoordinatorMock,
		syncTimer:               syncTimerMock,
		validatorGroupSelector:  validatorGroupSelector,
		epochStartNotifier:      epochStartSubscriber,
		antifloodHandler:        antifloodHandler,
		peerHonestyHandler:      peerHonestyHandler,
		headerSigVerifier:       headerSigVerifier,
		fallbackHeaderValidator: fallbackHeaderValidator,
		nodeRedundancyHandler:   nodeRedundancyHandler,
	}

	return container
}
