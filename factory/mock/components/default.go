package components

import (
	"time"

	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/factory/mock"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/testscommon"
	"github.com/ElrondNetwork/elrond-go/testscommon/cryptoMocks"
	dataRetrieverMock "github.com/ElrondNetwork/elrond-go/testscommon/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/testscommon/economicsmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/nodeTypeProviderMock"
	"github.com/ElrondNetwork/elrond-go/testscommon/p2pmocks"
	"github.com/ElrondNetwork/elrond-go/testscommon/shardingMocks"
	stateMock "github.com/ElrondNetwork/elrond-go/testscommon/state"
	statusHandlerMock "github.com/ElrondNetwork/elrond-go/testscommon/statusHandler"
	trieFactory "github.com/ElrondNetwork/elrond-go/trie/factory"
)

// GetDefaultCoreComponents -
func GetDefaultCoreComponents() *mock.CoreComponentsMock {
	return &mock.CoreComponentsMock{
		IntMarsh:            &testscommon.MarshalizerMock{},
		TxMarsh:             &testscommon.MarshalizerMock{},
		VmMarsh:             &testscommon.MarshalizerMock{},
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
		AppStatusHdl:          &statusHandlerMock.AppStatusHandlerStub{},
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
	}
}

// GetDefaultCryptoComponents -
func GetDefaultCryptoComponents() *mock.CryptoComponentsMock {
	return &mock.CryptoComponentsMock{
		PubKey:            &mock.PublicKeyMock{},
		PrivKey:           &mock.PrivateKeyStub{},
		PubKeyString:      "pubKey",
		PrivKeyBytes:      []byte("privKey"),
		PubKeyBytes:       []byte("pubKey"),
		BlockSig:          &mock.SinglesignMock{},
		TxSig:             &mock.SinglesignMock{},
		MultiSigContainer: &cryptoMocks.MultisignerStub{},
		PeerSignHandler:   &mock.PeerSignatureHandler{},
		BlKeyGen:          &mock.KeyGenMock{},
		TxKeyGen:          &mock.KeyGenMock{},
		MsgSigVerifier:    &testscommon.MessageSignVerifierMock{},
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
func GetDefaultStateComponents() *testscommon.StateComponentsMock {
	return &testscommon.StateComponentsMock{
		PeersAcc: &stateMock.AccountsStub{},
		Accounts: &stateMock.AccountsStub{},
		Tries:    &mock.TriesHolderStub{},
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
		Storage:           &mock.ChainStorerStub{},
		DataPool:          &dataRetrieverMock.PoolsHolderMock{},
		MiniBlockProvider: &mock.MiniBlocksProviderStub{},
	}
}

// GetDefaultProcessComponents -
func GetDefaultProcessComponents(shardCoordinator sharding.Coordinator) *mock.ProcessComponentsMock {
	return &mock.ProcessComponentsMock{
		NodesCoord:               &shardingMocks.NodesCoordinatorMock{},
		ShardCoord:               shardCoordinator,
		IntContainer:             &testscommon.InterceptorsContainerStub{},
		ResFinder:                &mock.ResolversFinderStub{},
		RoundHandlerField:        &testscommon.RoundHandlerMock{},
		EpochTrigger:             &testscommon.EpochStartTriggerStub{},
		EpochNotifier:            &mock.EpochStartNotifierStub{},
		ForkDetect:               &mock.ForkDetectorMock{},
		BlockProcess:             &mock.BlockProcessorStub{},
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
