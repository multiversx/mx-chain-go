package node

import (
	"math/big"
	"time"

	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/core/indexer"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/state"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/epochStart"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/ntp"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// WithMessenger sets up the messenger option for the Node
func WithMessenger(mes P2PMessenger) Option {
	return func(n *Node) error {
		if mes == nil || mes.IsInterfaceNil() {
			return ErrNilMessenger
		}
		n.messenger = mes
		return nil
	}
}

// WithMarshalizer sets up the marshalizer option for the Node
func WithMarshalizer(marshalizer marshal.Marshalizer, sizeCheckDelta uint32) Option {
	return func(n *Node) error {
		if marshalizer == nil || marshalizer.IsInterfaceNil() {
			return ErrNilMarshalizer
		}
		n.sizeCheckDelta = sizeCheckDelta
		n.marshalizer = marshalizer
		return nil
	}
}

// WithHasher sets up the hasher option for the Node
func WithHasher(hasher hashing.Hasher) Option {
	return func(n *Node) error {
		if hasher == nil || hasher.IsInterfaceNil() {
			return ErrNilHasher
		}
		n.hasher = hasher
		return nil
	}
}

// WithTxFeeHandler sets up the tx fee handler for the Node
func WithTxFeeHandler(feeHandler process.FeeHandler) Option {
	return func(n *Node) error {
		if feeHandler == nil || feeHandler.IsInterfaceNil() {
			return ErrNilTxFeeHandler
		}
		n.feeHandler = feeHandler
		return nil
	}
}

// WithAccountsAdapter sets up the accounts adapter option for the Node
func WithAccountsAdapter(accounts state.AccountsAdapter) Option {
	return func(n *Node) error {
		if accounts == nil || accounts.IsInterfaceNil() {
			return ErrNilAccountsAdapter
		}
		n.accounts = accounts
		return nil
	}
}

// WithAddressConverter sets up the address converter adapter option for the Node
func WithAddressConverter(addrConverter state.AddressConverter) Option {
	return func(n *Node) error {
		if addrConverter == nil || addrConverter.IsInterfaceNil() {
			return ErrNilAddressConverter
		}
		n.addrConverter = addrConverter
		return nil
	}
}

// WithBlockChain sets up the blockchain option for the Node
func WithBlockChain(blkc data.ChainHandler) Option {
	return func(n *Node) error {
		if blkc == nil || blkc.IsInterfaceNil() {
			return ErrNilBlockchain
		}
		n.blkc = blkc
		return nil
	}
}

// WithDataStore sets up the storage options for the Node
func WithDataStore(store dataRetriever.StorageService) Option {
	return func(n *Node) error {
		if store == nil || store.IsInterfaceNil() {
			return ErrNilStore
		}
		n.store = store
		return nil
	}
}

// WithTxSignPrivKey sets up the single sign private key option for the Node
func WithTxSignPrivKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil || sk.IsInterfaceNil() {
			return ErrNilPrivateKey
		}
		n.txSignPrivKey = sk
		return nil
	}
}

// WithPubKey sets up the multi sign pub key option for the Node
func WithPubKey(pk crypto.PublicKey) Option {
	return func(n *Node) error {
		if pk == nil || pk.IsInterfaceNil() {
			return ErrNilPublicKey
		}
		n.pubKey = pk
		return nil
	}
}

// WithPrivKey sets up the multi sign private key option for the Node
func WithPrivKey(sk crypto.PrivateKey) Option {
	return func(n *Node) error {
		if sk == nil || sk.IsInterfaceNil() {
			return ErrNilPrivateKey
		}
		n.privKey = sk
		return nil
	}
}

// WithKeyGen sets up the single sign key generator option for the Node
func WithKeyGen(keyGen crypto.KeyGenerator) Option {
	return func(n *Node) error {
		if keyGen == nil || keyGen.IsInterfaceNil() {
			return ErrNilSingleSignKeyGen
		}
		n.keyGen = keyGen
		return nil
	}
}

// WithKeyGenForAccounts sets up the balances key generator option for the Node
func WithKeyGenForAccounts(keyGenForAccounts crypto.KeyGenerator) Option {
	return func(n *Node) error {
		if keyGenForAccounts == nil || keyGenForAccounts.IsInterfaceNil() {
			return ErrNilKeyGenForBalances
		}
		n.keyGenForAccounts = keyGenForAccounts
		return nil
	}
}

// WithInitialNodesPubKeys sets up the initial nodes public key option for the Node
func WithInitialNodesPubKeys(pubKeys map[uint32][]string) Option {
	return func(n *Node) error {
		n.initialNodesPubkeys = pubKeys
		return nil
	}
}

// WithTxSignPubKey sets up the single sign public key option for the Node
func WithTxSignPubKey(pk crypto.PublicKey) Option {
	return func(n *Node) error {
		if pk == nil || pk.IsInterfaceNil() {
			return ErrNilPublicKey
		}

		n.txSignPubKey = pk
		return nil
	}
}

// WithRoundDuration sets up the round duration option for the Node
func WithRoundDuration(roundDuration uint64) Option {
	return func(n *Node) error {
		if roundDuration == 0 {
			return ErrZeroRoundDurationNotSupported
		}
		n.roundDuration = roundDuration
		return nil
	}
}

// WithConsensusGroupSize sets up the consensus group size option for the Node
func WithConsensusGroupSize(consensusGroupSize int) Option {
	return func(n *Node) error {
		if consensusGroupSize < 1 {
			return ErrNegativeOrZeroConsensusGroupSize
		}
		n.consensusGroupSize = consensusGroupSize
		return nil
	}
}

// WithSyncer sets up the syncTimer option for the Node
func WithSyncer(syncer ntp.SyncTimer) Option {
	return func(n *Node) error {
		if syncer == nil || syncer.IsInterfaceNil() {
			return ErrNilSyncTimer
		}
		n.syncTimer = syncer
		return nil
	}
}

// WithRounder sets up the rounder option for the Node
func WithRounder(rounder consensus.Rounder) Option {
	return func(n *Node) error {
		if rounder == nil || rounder.IsInterfaceNil() {
			return ErrNilRounder
		}
		n.rounder = rounder
		return nil
	}
}

// WithBlockProcessor sets up the block processor option for the Node
func WithBlockProcessor(blockProcessor process.BlockProcessor) Option {
	return func(n *Node) error {
		if blockProcessor == nil || blockProcessor.IsInterfaceNil() {
			return ErrNilBlockProcessor
		}
		n.blockProcessor = blockProcessor
		return nil
	}
}

// WithGenesisTime sets up the genesis time option for the Node
func WithGenesisTime(genesisTime time.Time) Option {
	return func(n *Node) error {
		n.genesisTime = genesisTime
		return nil
	}
}

// WithDataPool sets up the data pools option for the Node
func WithDataPool(dataPool dataRetriever.PoolsHolder) Option {
	return func(n *Node) error {
		if dataPool == nil || dataPool.IsInterfaceNil() {
			return ErrNilDataPool
		}
		n.dataPool = dataPool
		return nil
	}
}

// WithMetaDataPool sets up the data pools option for the Node
func WithMetaDataPool(dataPool dataRetriever.MetaPoolsHolder) Option {
	return func(n *Node) error {
		if dataPool == nil || dataPool.IsInterfaceNil() {
			return ErrNilDataPool
		}
		n.metaDataPool = dataPool
		return nil
	}
}

// WithShardCoordinator sets up the shard coordinator for the Node
func WithShardCoordinator(shardCoordinator sharding.Coordinator) Option {
	return func(n *Node) error {
		if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
			return ErrNilShardCoordinator
		}
		n.shardCoordinator = shardCoordinator
		return nil
	}
}

// WithNodesCoordinator sets up the nodes coordinator
func WithNodesCoordinator(nodesCoordinator sharding.NodesCoordinator) Option {
	return func(n *Node) error {
		if nodesCoordinator == nil {
			return ErrNilNodesCoordinator
		}
		n.nodesCoordinator = nodesCoordinator
		return nil
	}
}

// WithUint64ByteSliceConverter sets up the uint64 <-> []byte converter
func WithUint64ByteSliceConverter(converter typeConverters.Uint64ByteSliceConverter) Option {
	return func(n *Node) error {
		if converter == nil || converter.IsInterfaceNil() {
			return ErrNilUint64ByteSliceConverter
		}
		n.uint64ByteSliceConverter = converter
		return nil
	}
}

// WithInitialNodesBalances sets up the initial map of nodes public keys and their respective balances
func WithInitialNodesBalances(balances map[string]*big.Int) Option {
	return func(n *Node) error {
		if balances == nil {
			return ErrNilBalances
		}
		n.initialNodesBalances = balances
		return nil
	}
}

// WithSingleSigner sets up a singleSigner option for the Node
func WithSingleSigner(singleSigner crypto.SingleSigner) Option {
	return func(n *Node) error {
		if singleSigner == nil || singleSigner.IsInterfaceNil() {
			return ErrNilSingleSig
		}
		n.singleSigner = singleSigner
		return nil
	}
}

// WithTxSingleSigner sets up a txSingleSigner option for the Node
func WithTxSingleSigner(txSingleSigner crypto.SingleSigner) Option {
	return func(n *Node) error {
		if txSingleSigner == nil || txSingleSigner.IsInterfaceNil() {
			return ErrNilSingleSig
		}
		n.txSingleSigner = txSingleSigner
		return nil
	}
}

// WithMultiSigner sets up the multiSigner option for the Node
func WithMultiSigner(multiSigner crypto.MultiSigner) Option {
	return func(n *Node) error {
		if multiSigner == nil || multiSigner.IsInterfaceNil() {
			return ErrNilMultiSig
		}
		n.multiSigner = multiSigner
		return nil
	}
}

// WithForkDetector sets up the multiSigner option for the Node
func WithForkDetector(forkDetector process.ForkDetector) Option {
	return func(n *Node) error {
		if forkDetector == nil || forkDetector.IsInterfaceNil() {
			return ErrNilForkDetector
		}
		n.forkDetector = forkDetector
		return nil
	}
}

// WithInterceptorsContainer sets up the interceptors container option for the Node
func WithInterceptorsContainer(interceptorsContainer process.InterceptorsContainer) Option {
	return func(n *Node) error {
		if interceptorsContainer == nil || interceptorsContainer.IsInterfaceNil() {
			return ErrNilInterceptorsContainer
		}
		n.interceptorsContainer = interceptorsContainer
		return nil
	}
}

// WithResolversFinder sets up the resolvers finder option for the Node
func WithResolversFinder(resolversFinder dataRetriever.ResolversFinder) Option {
	return func(n *Node) error {
		if resolversFinder == nil || resolversFinder.IsInterfaceNil() {
			return ErrNilResolversFinder
		}
		n.resolversFinder = resolversFinder
		return nil
	}
}

// WithConsensusType sets up the consensus type option for the Node
func WithConsensusType(consensusType string) Option {
	return func(n *Node) error {
		n.consensusType = consensusType
		return nil
	}
}

// WithTxStorageSize sets up a txStorageSize option for the Node
func WithTxStorageSize(txStorageSize uint32) Option {
	return func(n *Node) error {
		n.txStorageSize = txStorageSize
		return nil
	}
}

// WithBootstrapRoundIndex sets up a bootstrapRoundIndex option for the Node
func WithBootstrapRoundIndex(bootstrapRoundIndex uint64) Option {
	return func(n *Node) error {
		n.bootstrapRoundIndex = bootstrapRoundIndex
		return nil
	}
}

// WithEpochStartTrigger sets up an start of epoch trigger option for the node
func WithEpochStartTrigger(epochStartTrigger epochStart.TriggerHandler) Option {
	return func(n *Node) error {
		if check.IfNil(epochStartTrigger) {
			return ErrNilEpochStartTrigger
		}
		n.epochStartTrigger = epochStartTrigger
		return nil
	}
}

// WithAppStatusHandler sets up which handler will monitor the status of the node
func WithAppStatusHandler(aph core.AppStatusHandler) Option {
	return func(n *Node) error {
		if aph == nil || aph.IsInterfaceNil() {
			return ErrNilStatusHandler

		}
		n.appStatusHandler = aph
		return nil
	}
}

// WithIndexer sets up a indexer for the Node
func WithIndexer(indexer indexer.Indexer) Option {
	return func(n *Node) error {
		n.indexer = indexer
		return nil
	}
}

// WithBlackListHandler sets up a black list handler for the Node
func WithBlackListHandler(blackListHandler process.BlackListHandler) Option {
	return func(n *Node) error {
		if check.IfNil(blackListHandler) {
			return ErrNilBlackListHandler
		}
		n.blackListHandler = blackListHandler
		return nil
	}
}

// WithBootStorer sets up a boot storer for the Node
func WithBootStorer(bootStorer process.BootStorer) Option {
	return func(n *Node) error {
		if check.IfNil(bootStorer) {
			return ErrNilBootStorer
		}
		n.bootStorer = bootStorer
		return nil
	}
}

// WithRequestedItemsHandler sets up a requested items handler for the Node
func WithRequestedItemsHandler(requestedItemsHandler dataRetriever.RequestedItemsHandler) Option {
	return func(n *Node) error {
		if check.IfNil(requestedItemsHandler) {
			return ErrNilRequestedItemsHandler
		}
		n.requestedItemsHandler = requestedItemsHandler
		return nil
	}
}

// WithHeaderSigVerifier sets up a header sig verifier for the Node
func WithHeaderSigVerifier(headerSigVerifier spos.RandSeedVerifier) Option {
	return func(n *Node) error {
		if check.IfNil(headerSigVerifier) {
			return ErrNilHeaderSigVerifier
		}
		n.headerSigVerifier = headerSigVerifier
		return nil
	}
}

// WithValidatorStatistics sets up the validator statistics fro the node
func WithValidatorStatistics(validatorStatistics process.ValidatorStatisticsProcessor) Option {
	return func(n *Node) error {
		if check.IfNil(validatorStatistics) {
			return ErrNilValidatorStatistics
		}
		n.validatorStatistics = validatorStatistics
		return nil
	}
}

// WithChainID sets up the chain ID on which the current node is supposed to work on
func WithChainID(chainID []byte) Option {
	return func(n *Node) error {
		if len(chainID) == 0 {
			return ErrInvalidChainID
		}
		n.chainID = chainID

		return nil
	}
}
