package factory

import (
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/common"
	"github.com/ElrondNetwork/elrond-go/config"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/epochStart/bootstrap/disabled"
	disabledGenesis "github.com/ElrondNetwork/elrond-go/genesis/process/disabled"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/factory/interceptorscontainer"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage/timecache"
	"github.com/ElrondNetwork/elrond-go/update"
)

const timeSpanForBadHeaders = time.Minute

// ArgsEpochStartInterceptorContainer holds the arguments needed for creating a new epoch start interceptors
// container factory
type ArgsEpochStartInterceptorContainer struct {
	CoreComponents            process.CoreComponentsHolder
	CryptoComponents          process.CryptoComponentsHolder
	Config                    config.Config
	ShardCoordinator          sharding.Coordinator
	Messenger                 process.TopicHandler
	DataPool                  dataRetriever.PoolsHolder
	WhiteListHandler          update.WhiteListHandler
	WhiteListerVerifiedTxs    update.WhiteListHandler
	AddressPubkeyConv         core.PubkeyConverter
	NonceConverter            typeConverters.Uint64ByteSliceConverter
	ChainID                   []byte
	ArgumentsParser           process.ArgumentsParser
	HeaderIntegrityVerifier   process.HeaderIntegrityVerifier
	EnableSignTxWithHashEpoch uint32
	EpochNotifier             process.EpochNotifier
	RequestHandler            process.RequestHandler
}

// NewEpochStartInterceptorsContainer will return a real interceptors container factory, but with many disabled components
func NewEpochStartInterceptorsContainer(args ArgsEpochStartInterceptorContainer) (process.InterceptorsContainer, error) {
	if check.IfNil(args.CoreComponents) {
		return nil, epochStart.ErrNilCoreComponentsHolder
	}
	if check.IfNil(args.CryptoComponents) {
		return nil, epochStart.ErrNilCryptoComponentsHolder
	}

	if check.IfNil(args.CoreComponents.AddressPubKeyConverter()) {
		return nil, epochStart.ErrNilPubkeyConverter
	}

	cryptoComponents := args.CryptoComponents.Clone().(process.CryptoComponentsHolder)
	err := cryptoComponents.SetMultiSigner(disabled.NewMultiSigner())
	if err != nil {
		return nil, err
	}

	nodesCoordinator := disabled.NewNodesCoordinator()
	storer := disabled.NewChainStorer()
	antiFloodHandler := disabled.NewAntiFloodHandler()
	accountsAdapter := disabled.NewAccountsAdapter()
	blackListHandler := timecache.NewTimeCache(timeSpanForBadHeaders)
	feeHandler := &disabledGenesis.FeeHandler{}
	headerSigVerifier := disabled.NewHeaderSigVerifier()
	sizeCheckDelta := 0
	validityAttester := disabled.NewValidityAttester()
	epochStartTrigger := disabled.NewEpochStartTrigger()

	containerFactoryArgs := interceptorscontainer.CommonInterceptorsContainerFactoryArgs{
		CoreComponents:            args.CoreComponents,
		CryptoComponents:          cryptoComponents,
		ShardCoordinator:          args.ShardCoordinator,
		NodesCoordinator:          nodesCoordinator,
		Messenger:                 args.Messenger,
		Store:                     storer,
		DataPool:                  args.DataPool,
		Accounts:                  accountsAdapter,
		MaxTxNonceDeltaAllowed:    common.MaxTxNonceDeltaAllowed,
		TxFeeHandler:              feeHandler,
		BlockBlackList:            blackListHandler,
		HeaderSigVerifier:         headerSigVerifier,
		HeaderIntegrityVerifier:   args.HeaderIntegrityVerifier,
		SizeCheckDelta:            uint32(sizeCheckDelta),
		ValidityAttester:          validityAttester,
		EpochStartTrigger:         epochStartTrigger,
		WhiteListHandler:          args.WhiteListHandler,
		WhiteListerVerifiedTxs:    args.WhiteListerVerifiedTxs,
		AntifloodHandler:          antiFloodHandler,
		ArgumentsParser:           args.ArgumentsParser,
		EnableSignTxWithHashEpoch: args.EnableSignTxWithHashEpoch,
		PreferredPeersHolder:      disabled.NewPreferredPeersHolder(),
		RequestHandler:            args.RequestHandler,
	}

	interceptorsContainerFactory, err := interceptorscontainer.NewMetaInterceptorsContainerFactory(containerFactoryArgs)
	if err != nil {
		return nil, err
	}

	container, err := interceptorsContainerFactory.Create()
	if err != nil {
		return nil, err
	}

	err = interceptorsContainerFactory.AddShardTrieNodeInterceptors(container)
	if err != nil {
		return nil, err
	}

	return container, nil
}
