package mock

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-core-go/marshal"
	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
	"github.com/multiversx/mx-chain-go/testscommon"
	consensusMocks "github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	"github.com/multiversx/mx-chain-go/testscommon/hashingMocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
)

// InitChronologyHandlerMock -
func InitChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &ChronologyHandlerMock{}
	return chr
}

// InitBlockProcessorMock -
func InitBlockProcessorMock(marshaller marshal.Marshalizer) *testscommon.BlockProcessorStub {
	blockProcessorMock := &testscommon.BlockProcessorStub{}
	blockProcessorMock.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		emptyBlock := &block.Body{}
		_ = header.SetRootHash([]byte{})
		return header, emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertCurrentBlockCalled = func() {}
	blockProcessorMock.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
		return header, body, nil
	}
	blockProcessorMock.DecodeBlockBodyCalled = func(dta []byte) data.BodyHandler {
		return &block.Body{}
	}
	blockProcessorMock.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
		header := &block.Header{}
		_ = marshaller.Unmarshal(header, dta)

		return header
	}
	blockProcessorMock.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	blockProcessorMock.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
		return &block.Header{
			Round: round,
			Nonce: nonce,
		}, nil
	}

	return blockProcessorMock
}

// InitBlockProcessorHeaderV2Mock -
func InitBlockProcessorHeaderV2Mock() *testscommon.BlockProcessorStub {
	blockProcessorMock := &testscommon.BlockProcessorStub{}
	blockProcessorMock.CreateBlockCalled = func(header data.HeaderHandler, haveTime func() bool) (data.HeaderHandler, data.BodyHandler, error) {
		emptyBlock := &block.Body{}
		_ = header.SetRootHash([]byte{})
		return header, emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertCurrentBlockCalled = func() {}
	blockProcessorMock.ProcessBlockCalled = func(header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) (data.HeaderHandler, data.BodyHandler, error) {
		return header, body, nil
	}
	blockProcessorMock.DecodeBlockBodyCalled = func(dta []byte) data.BodyHandler {
		return &block.Body{}
	}
	blockProcessorMock.DecodeBlockHeaderCalled = func(dta []byte) data.HeaderHandler {
		return &block.HeaderV2{
			Header:            &block.Header{},
			ScheduledRootHash: []byte{},
		}
	}
	blockProcessorMock.MarshalizedDataToBroadcastCalled = func(header data.HeaderHandler, body data.BodyHandler) (map[uint32][]byte, map[string][][]byte, error) {
		return make(map[uint32][]byte), make(map[string][][]byte), nil
	}
	blockProcessorMock.CreateNewHeaderCalled = func(round uint64, nonce uint64) (data.HeaderHandler, error) {
		return &block.HeaderV2{
			Header: &block.Header{

				Round: round,
				Nonce: nonce,
			},
			ScheduledRootHash: []byte{},
		}, nil
	}

	return blockProcessorMock
}

// InitMultiSignerMock -
func InitMultiSignerMock() *cryptoMocks.MultisignerMock {
	multiSigner := cryptoMocks.NewMultiSigner()
	multiSigner.VerifySignatureShareCalled = func(publicKey []byte, message []byte, sig []byte) error {
		return nil
	}
	multiSigner.VerifyAggregatedSigCalled = func(pubKeysSigners [][]byte, message []byte, aggSig []byte) error {
		return nil
	}
	multiSigner.AggregateSigsCalled = func(pubKeysSigners [][]byte, signatures [][]byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}
	multiSigner.CreateSignatureShareCalled = func(privateKeyBytes []byte, message []byte) ([]byte, error) {
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

// InitConsensusCoreHeaderV2 -
func InitConsensusCoreHeaderV2() *ConsensusCoreMock {
	consensusCoreMock := InitConsensusCore()
	consensusCoreMock.blockProcessor = InitBlockProcessorHeaderV2Mock()

	return consensusCoreMock
}

// InitConsensusCore -
func InitConsensusCore() *ConsensusCoreMock {
	multiSignerMock := InitMultiSignerMock()

	return InitConsensusCoreWithMultiSigner(multiSignerMock)
}

// InitConsensusCoreWithMultiSigner -
func InitConsensusCoreWithMultiSigner(multiSigner crypto.MultiSigner) *ConsensusCoreMock {
	blockChain := &testscommon.ChainHandlerStub{
		GetGenesisHeaderCalled: func() data.HeaderHandler {
			return &block.Header{}
		},
	}
	marshalizerMock := MarshalizerMock{}
	blockProcessorMock := InitBlockProcessorMock(marshalizerMock)
	bootstrapperMock := &BootstrapperStub{}
	broadcastMessengerMock := &BroadcastMessengerMock{
		BroadcastConsensusMessageCalled: func(message *consensus.Message) error {
			return nil
		},
	}

	chronologyHandlerMock := InitChronologyHandlerMock()
	hasherMock := &hashingMocks.HasherMock{}
	roundHandlerMock := &RoundHandlerMock{}
	shardCoordinatorMock := ShardCoordinatorMock{}
	syncTimerMock := &SyncTimerMock{}
	validatorGroupSelector := &shardingMocks.NodesCoordinatorMock{
		ComputeValidatorsGroupCalled: func(randomness []byte, round uint64, shardId uint32, epoch uint32) ([]nodesCoordinator.Validator, error) {
			defaultSelectionChances := uint32(1)
			return []nodesCoordinator.Validator{
				shardingMocks.NewValidatorMock([]byte("A"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("B"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("C"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("D"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("E"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("F"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("G"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("H"), 1, defaultSelectionChances),
				shardingMocks.NewValidatorMock([]byte("I"), 1, defaultSelectionChances),
			}, nil
		},
	}
	epochStartSubscriber := &EpochStartNotifierStub{}
	antifloodHandler := &P2PAntifloodHandlerStub{}
	headerPoolSubscriber := &testscommon.HeadersCacherStub{}
	peerHonestyHandler := &testscommon.PeerHonestyHandlerStub{}
	headerSigVerifier := &HeaderSigVerifierStub{}
	fallbackHeaderValidator := &testscommon.FallBackHeaderValidatorStub{}
	nodeRedundancyHandler := &NodeRedundancyHandlerStub{}
	scheduledProcessor := &consensusMocks.ScheduledProcessorStub{}
	messageSigningHandler := &MessageSigningHandlerStub{}
	peerBlacklistHandler := &PeerBlacklistHandlerStub{}
	multiSignerContainer := cryptoMocks.NewMultiSignerContainerMock(multiSigner)
	signingHandler := &consensusMocks.SigningHandlerStub{}

	container := &ConsensusCoreMock{
		blockChain:              blockChain,
		blockProcessor:          blockProcessorMock,
		headersSubscriber:       headerPoolSubscriber,
		bootstrapper:            bootstrapperMock,
		broadcastMessenger:      broadcastMessengerMock,
		chronologyHandler:       chronologyHandlerMock,
		hasher:                  hasherMock,
		marshalizer:             marshalizerMock,
		multiSignerContainer:    multiSignerContainer,
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
		scheduledProcessor:      scheduledProcessor,
		messageSigningHandler:   messageSigningHandler,
		peerBlacklistHandler:    peerBlacklistHandler,
		signingHandler:          signingHandler,
	}

	return container
}
