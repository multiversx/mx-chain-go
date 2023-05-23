package components

import (
	"time"

	crypto "github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/factory/mock"
	"github.com/multiversx/mx-chain-go/sharding"
	"github.com/multiversx/mx-chain-go/testscommon"
	"github.com/multiversx/mx-chain-go/testscommon/consensus"
	"github.com/multiversx/mx-chain-go/testscommon/cryptoMocks"
	dataRetrieverTests "github.com/multiversx/mx-chain-go/testscommon/dataRetriever"
	"github.com/multiversx/mx-chain-go/testscommon/economicsmocks"
	epochNotifierMock "github.com/multiversx/mx-chain-go/testscommon/epochNotifier"
	"github.com/multiversx/mx-chain-go/testscommon/factory"
	"github.com/multiversx/mx-chain-go/testscommon/marshallerMock"
	"github.com/multiversx/mx-chain-go/testscommon/nodeTypeProviderMock"
	"github.com/multiversx/mx-chain-go/testscommon/p2pmocks"
	"github.com/multiversx/mx-chain-go/testscommon/shardingMocks"
	stateMock "github.com/multiversx/mx-chain-go/testscommon/state"
	"github.com/multiversx/mx-chain-go/testscommon/storage"
	trieMock "github.com/multiversx/mx-chain-go/testscommon/trie"
	trieFactory "github.com/multiversx/mx-chain-go/trie/factory"
)

// GetDefaultCoreComponents -
func GetDefaultCoreComponents() *mock.CoreComponentsMock {
	return &mock.CoreComponentsMock{
		IntMarsh:            &marshallerMock.MarshalizerMock{},
		TxMarsh:             &marshallerMock.MarshalizerMock{},
		VmMarsh:             &marshallerMock.MarshalizerMock{},
		Hash:                &testscommon.HasherStub{},
		UInt64ByteSliceConv: testscommon.NewNonceHashConverterMock(),
		AddrPubKeyConv:      testscommon.NewPubkeyConverterMock(32),
		ValPubKeyConv:       testscommon.NewPubkeyConverterMock(32),
		PathHdl:             &testscommon.PathManagerStub{},
		ChainIdCalled: func() string {
			return "chainID"
		},
		MinTransactionVersionCalled: func() uint32 {
			return 1
		},
		WatchdogTimer:         &testscommon.WatchdogMock{},
		AlarmSch:              &testscommon.AlarmSchedulerStub{},
		NtpSyncTimer:          &testscommon.SyncTimerStub{},
		RoundHandlerField:     &testscommon.RoundHandlerMock{},
		EconomicsHandler:      &economicsmocks.EconomicsHandlerStub{},
		RatingsConfig:         &testscommon.RatingsInfoMock{},
		RatingHandler:         &testscommon.RaterMock{},
		NodesConfig:           &testscommon.NodesSetupStub{},
		StartTime:             time.Time{},
		NodeTypeProviderField: &nodeTypeProviderMock.NodeTypeProviderStub{},
		EpochChangeNotifier:   &epochNotifierMock.EpochNotifierStub{},
	}
}

// GetDefaultCryptoComponents -
func GetDefaultCryptoComponents() *mock.CryptoComponentsMock {
	return &mock.CryptoComponentsMock{
		PubKey:            &mock.PublicKeyMock{},
		PrivKey:           &mock.PrivateKeyStub{},
		P2pPubKey:         &mock.PublicKeyMock{},
		P2pPrivKey:        mock.NewP2pPrivateKeyMock(),
		P2pSig:            &mock.SinglesignMock{},
		PubKeyString:      "pubKey",
		PubKeyBytes:       []byte("pubKey"),
		BlockSig:          &mock.SinglesignMock{},
		TxSig:             &mock.SinglesignMock{},
		MultiSigContainer: cryptoMocks.NewMultiSignerContainerMock(&cryptoMocks.MultisignerMock{}),
		PeerSignHandler:   &mock.PeerSignatureHandler{},
		BlKeyGen:          &mock.KeyGenMock{},
		TxKeyGen:          &mock.KeyGenMock{},
		P2PKeyGen:         &mock.KeyGenMock{},
		MsgSigVerifier:    &testscommon.MessageSignVerifierMock{},
		SigHandler:        &consensus.SigningHandlerStub{},
	}
}

// GetDefaultNetworkComponents -
func GetDefaultNetworkComponents() *mock.NetworkComponentsMock {
	return &mock.NetworkComponentsMock{
		Messenger:       &p2pmocks.MessengerStub{},
		InputAntiFlood:  &mock.P2PAntifloodHandlerStub{},
		OutputAntiFlood: &mock.P2PAntifloodHandlerStub{},
		PeerBlackList:   &mock.PeerBlackListHandlerStub{},
	}
}

// GetDefaultStateComponents -
func GetDefaultStateComponents() *factory.StateComponentsMock {
	return &factory.StateComponentsMock{
		PeersAcc: &stateMock.AccountsStub{},
		Accounts: &stateMock.AccountsStub{},
		Tries:    &trieMock.TriesHolderStub{},
		StorageManagers: map[string]common.StorageManager{
			"0":                         &testscommon.StorageManagerStub{},
			trieFactory.UserAccountTrie: &testscommon.StorageManagerStub{},
			trieFactory.PeerAccountTrie: &testscommon.StorageManagerStub{},
		},
	}
}

// GetDefaultDataComponents -
func GetDefaultDataComponents() *mock.DataComponentsMock {
	return &mock.DataComponentsMock{
		Blkc:              &testscommon.ChainHandlerStub{},
		Storage:           &storage.ChainStorerStub{},
		DataPool:          &dataRetrieverTests.PoolsHolderMock{},
		MiniBlockProvider: &mock.MiniBlocksProviderStub{},
	}
}

// GetDefaultProcessComponents -
func GetDefaultProcessComponents(shardCoordinator sharding.Coordinator) *mock.ProcessComponentsMock {
	return &mock.ProcessComponentsMock{
		NodesCoord:               &shardingMocks.NodesCoordinatorMock{},
		ShardCoord:               shardCoordinator,
		IntContainer:             &testscommon.InterceptorsContainerStub{},
		ResContainer:             &dataRetrieverTests.ResolversContainerStub{},
		ReqFinder:                &dataRetrieverTests.RequestersFinderStub{},
		RoundHandlerField:        &testscommon.RoundHandlerMock{},
		EpochTrigger:             &testscommon.EpochStartTriggerStub{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		ForkDetect:               &mock.ForkDetectorMock{},
		BlockProcess:             &testscommon.BlockProcessorStub{},
		BlackListHdl:             &testscommon.TimeCacheStub{},
		BootSore:                 &mock.BootstrapStorerMock{},
		HeaderSigVerif:           &mock.HeaderSigVerifierStub{},
		HeaderIntegrVerif:        &mock.HeaderIntegrityVerifierStub{},
		ValidatorStatistics:      &mock.ValidatorStatisticsProcessorStub{},
		ValidatorProvider:        &mock.ValidatorsProviderStub{},
		BlockTrack:               &mock.BlockTrackerStub{},
		PendingMiniBlocksHdl:     &mock.PendingMiniBlocksHandlerStub{},
		ReqHandler:               &testscommon.RequestHandlerStub{},
		TxLogsProcess:            &mock.TxLogProcessorMock{},
		HeaderConstructValidator: &mock.HeaderValidatorStub{},
		PeerMapper:               &p2pmocks.NetworkShardingCollectorStub{},
		FallbackHdrValidator:     &testscommon.FallBackHeaderValidatorStub{},
		NodeRedundancyHandlerInternal: &mock.RedundancyHandlerStub{
			IsRedundancyNodeCalled: func() bool {
				return false
			},
			IsMainMachineActiveCalled: func() bool {
				return false
			},
			ObserverPrivateKeyCalled: func() crypto.PrivateKey {
				return &mock.PrivateKeyStub{}
			},
		},
		HardforkTriggerField: &testscommon.HardforkTriggerStub{},
	}
}
