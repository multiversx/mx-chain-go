package errors

import (
	"errors"
)

// ErrAccountsAdapterCreation signals that the accounts adapter cannot be created based on provided data
var ErrAccountsAdapterCreation = errors.New("error creating accounts adapter")

// ErrBlockchainCreation signals that the blockchain cannot be created
var ErrBlockchainCreation = errors.New("can not create blockchain")

// ErrDataPoolCreation signals that the data pool cannot be created
var ErrDataPoolCreation = errors.New("can not create data pool")

// ErrDataStoreCreation signals that the data store cannot be created
var ErrDataStoreCreation = errors.New("can not create data store")

// ErrGenesisBlockNotInitialized signals that genesis block is not initialized
var ErrGenesisBlockNotInitialized = errors.New("genesis block is not initialized")

// ErrHasherCreation signals that the hasher cannot be created based on provided data
var ErrHasherCreation = errors.New("error creating hasher")

// ErrInvalidChainID signals that an invalid chain ID has been provided
var ErrInvalidChainID = errors.New("invalid chain ID in consensus")

// ErrInvalidConsensusConfig signals that an invalid consensus type is specified in the configuration file
var ErrInvalidConsensusConfig = errors.New("invalid consensus type provided in config file")

// ErrInvalidRoundDuration signals that an invalid round duration has been provided
var ErrInvalidRoundDuration = errors.New("invalid round duration provided")

// ErrInvalidTransactionVersion signals  that an invalid transaction version has been provided
var ErrInvalidTransactionVersion = errors.New("invalid transaction version")

// ErrInvalidWorkingDir signals that an invalid working directory has been provided
var ErrInvalidWorkingDir = errors.New("invalid working directory")

// ErrMarshalizerCreation signals that the marshalizer cannot be created based on provided data
var ErrMarshalizerCreation = errors.New("error creating marshalizer")

// ErrMissingMultiHasherConfig signals that the multihasher type isn't specified in the configuration file
var ErrMissingMultiHasherConfig = errors.New("no multisig hasher provided in config file")

// ErrMultiSigHasherMissmatch signals that an invalid multisig hasher was provided
var ErrMultiSigHasherMissmatch = errors.New("wrong multisig hasher provided for bls consensus type")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("nil accounts adapter")

// ErrNilAccountsParser signals that a nil accounts parser has been provided
var ErrNilAccountsParser = errors.New("nil accounts parser")

// ErrNilAddressPublicKeyConverter signals that an operation has been attempted to or with a nil public key converter implementation
var ErrNilAddressPublicKeyConverter = errors.New("nil address pubkey converter")

// ErrNilAlarmScheduler is raised when a valid alarm scheduler is expected but nil is used
var ErrNilAlarmScheduler = errors.New("nil alarm scheduler")

// ErrNilBlackListHandler signals that a nil black list handler was provided
var ErrNilBlackListHandler = errors.New("nil black list handler")

// ErrNilBlockChainHandler is raised when a valid blockchain handler is expected but nil used
var ErrNilBlockChainHandler = errors.New("blockchain handler is nil")

// ErrNilBlockProcessor is raised when a valid block processor is expected but nil used
var ErrNilBlockProcessor = errors.New("block processor is nil")

// ErrNilBlockSigner signals the nil block signer was provided
var ErrNilBlockSigner = errors.New("nil block signer")

// ErrNilP2pSigner signals the nil p2p signer was provided
var ErrNilP2pSigner = errors.New("nil p2p single signer")

// ErrNilBlockSignKeyGen is raised when a valid block sign key generator is expected but nil used
var ErrNilBlockSignKeyGen = errors.New("block sign key generator is nil")

// ErrNilBlockTracker signals that a nil block tracker has been provided
var ErrNilBlockTracker = errors.New("trying to set nil block tracker")

// ErrNilBootStorer signals that the provided boot storer is nil
var ErrNilBootStorer = errors.New("nil boot storer")

// ErrNilBootstrapComponents signals that the provided instance of bootstrap components is nil
var ErrNilBootstrapComponents = errors.New("nil bootstrap components")

// ErrNilBootstrapComponentsHolder signals that the provided bootstrap components holder is nil
var ErrNilBootstrapComponentsHolder = errors.New("nil bootstrap components holder")

// ErrNilBootstrapComponentsFactory signals that the provided bootstrap components factory is nil
var ErrNilBootstrapComponentsFactory = errors.New("nil bootstrap components factory")

// ErrNilConsensusComponentsFactory signals that the provided consensus components factory is nil
var ErrNilConsensusComponentsFactory = errors.New("nil consensus components factory")

// ErrNilCryptoComponentsFactory signals that the provided crypto components factory is nil
var ErrNilCryptoComponentsFactory = errors.New("nil crypto components factory")

// ErrNilCoreComponentsFactory signals that the provided core components factory is nil
var ErrNilCoreComponentsFactory = errors.New("nil core components factory")

// ErrNilDataComponentsFactory signals that the provided data components factory is nil
var ErrNilDataComponentsFactory = errors.New("nil data components factory")

// ErrNilHeartbeatV2ComponentsFactory signals that the provided heartbeatV2 components factory is nil
var ErrNilHeartbeatV2ComponentsFactory = errors.New("nil heartbeatV2 components factory")

// ErrNilNetworkComponentsFactory signals that the provided network components factory is nil
var ErrNilNetworkComponentsFactory = errors.New("nil network components factory")

// ErrNilProcessComponentsFactory signals that the provided process components factory is nil
var ErrNilProcessComponentsFactory = errors.New("nil process components factory")

// ErrNilStateComponentsFactory signals that the provided state components factory is nil
var ErrNilStateComponentsFactory = errors.New("nil state components factory")

// ErrNilStatusComponentsFactory signals that the provided status components factory is nil
var ErrNilStatusComponentsFactory = errors.New("nil status components factory")

// ErrNilStatusCoreComponentsFactory signals that an operation has been attempted with nil status core components factory
var ErrNilStatusCoreComponentsFactory = errors.New("nil status core components factory provided")

// ErrNilRunTypeComponentsFactory signals that the provided runType components factory is nil
var ErrNilRunTypeComponentsFactory = errors.New("nil runType components factory")

// ErrNilBootstrapParamsHandler signals that the provided bootstrap parameters handler is nil
var ErrNilBootstrapParamsHandler = errors.New("nil bootstrap parameters handler")

// ErrNilBroadcastMessenger is raised when a valid broadcast messenger is expected but nil used
var ErrNilBroadcastMessenger = errors.New("broadcast messenger is nil")

// ErrNilChronologyHandler is raised when a valid chronology handler is expected but nil used
var ErrNilChronologyHandler = errors.New("chronology handler is nil")

// ErrNilConsensusComponentsHolder signals that a nil consensus components holder was provided
var ErrNilConsensusComponentsHolder = errors.New("nil consensus components holder")

// ErrNilConsensusWorker signals that a nil consensus worker was provided
var ErrNilConsensusWorker = errors.New("nil consensus worker")

// ErrNilCoreComponents signals that an operation has been attempted with nil core components
var ErrNilCoreComponents = errors.New("nil core components provided")

// ErrNilStatusCoreComponents signals that an operation has been attempted with nil status core components
var ErrNilStatusCoreComponents = errors.New("nil status core components provided")

// ErrNilCoreComponentsHolder signals that a nil core components holder was provided
var ErrNilCoreComponentsHolder = errors.New("nil core components holder")

// ErrNilCryptoComponents signals that a nil crypto components has been provided
var ErrNilCryptoComponents = errors.New("nil crypto components provided")

// ErrNilCryptoComponentsHolder signals that a nil crypto components holder was provided
var ErrNilCryptoComponentsHolder = errors.New("nil crypto components holder")

// ErrNilDataComponents signals that a nil data components instance was provided
var ErrNilDataComponents = errors.New("nil data components provided")

// ErrNilDataComponentsHolder signals that a nil data components holder has been provided
var ErrNilDataComponentsHolder = errors.New("nil data components holder")

// ErrNilEconomicsData signals that a nil economics data handler has been provided
var ErrNilEconomicsData = errors.New("nil economics data provided")

// ErrNilEconomicsHandler signals that a nil economics handler has been provided
var ErrNilEconomicsHandler = errors.New("nil economics handler")

// ErrNilOutportHandler signals that a nil outport handler has been provided
var ErrNilOutportHandler = errors.New("nil outport handler")

// ErrNilEpochNotifier signals that a nil epoch notifier has been provided
var ErrNilEpochNotifier = errors.New("nil epoch notifier")

// ErrNilEpochStartBootstrapper signals that a nil epoch start bootstrapper was provided
var ErrNilEpochStartBootstrapper = errors.New("nil epoch start bootstrapper")

// ErrNilEpochStartNotifier signals that a nil epoch start notifier was provided
var ErrNilEpochStartNotifier = errors.New("nil epoch start notifier provided")

// ErrNilEpochStartTrigger signals that a nil start of epoch trigger has been provided
var ErrNilEpochStartTrigger = errors.New("nil start of epoch trigger")

// ErrNilFallbackHeaderValidator signals that a nil fallback header validator has been provided
var ErrNilFallbackHeaderValidator = errors.New("nil fallback header validator")

// ErrNilForkDetector is raised when a valid fork detector is expected but nil used
var ErrNilForkDetector = errors.New("fork detector is nil")

// ErrNilGasSchedule signals that an operation has been attempted with a nil gas schedule
var ErrNilGasSchedule = errors.New("nil gas schedule")

// ErrNilHasher is raised when a valid hasher is expected but nil used
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilTxSignHasher is raised when a nil tx sign hasher is provided
var ErrNilTxSignHasher = errors.New("nil tx signing hasher")

// ErrNilHeaderConstructionValidator signals that a nil header construction validator was provided
var ErrNilHeaderConstructionValidator = errors.New("nil header construction validator")

// ErrNilHeaderIntegrityVerifier signals that a nil header integrity verifier has been provided
var ErrNilHeaderIntegrityVerifier = errors.New("nil header integrity verifier")

// ErrNilHeaderSigVerifier signals that a nil header sig verifier has been provided
var ErrNilHeaderSigVerifier = errors.New("")

// ErrNilHeartbeatV2Components signals that a nil heartbeatV2 components instance was provided
var ErrNilHeartbeatV2Components = errors.New("nil heartbeatV2 component")

// ErrNilHeartbeatV2Sender signals that a nil heartbeatV2 sender was provided
var ErrNilHeartbeatV2Sender = errors.New("nil heartbeatV2 sender")

// ErrNilInputAntiFloodHandler signals that a nil input antiflood handler was provided
var ErrNilInputAntiFloodHandler = errors.New("nil input antiflood handler")

// ErrNilInterceptorsContainer signals that a nil interceptors container was provided
var ErrNilInterceptorsContainer = errors.New("nil interceptors container")

// ErrNilInternalMarshalizer signals that a nil internal marshalizer was provided
var ErrNilInternalMarshalizer = errors.New("nil internal marshalizer")

// ErrNilKeyLoader signals that a nil key loader was provided
var ErrNilKeyLoader = errors.New("nil key loader")

// ErrNilMarshalizer signals that a nil marshalizer was provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilMessageSignVerifier signals that a nil message signiature verifier was provided
var ErrNilMessageSignVerifier = errors.New("nil message sign verifier")

// ErrNilMessenger signals that a nil messenger was provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilMiniBlocksProvider signals a nil miniBlocks provider
var ErrNilMiniBlocksProvider = errors.New("nil miniBlocks provider")

// ErrNilMultiSigner signals that a nil multi-signer was provided
var ErrNilMultiSigner = errors.New("nil multi signer")

// ErrNilNetworkComponents signals that a nil network components instance was provided
var ErrNilNetworkComponents = errors.New("nil network components")

// ErrNilNetworkComponentsHolder signals that a nil network components holder was provided
var ErrNilNetworkComponentsHolder = errors.New("nil network components holder")

// ErrNilNodesConfig signals that a nil nodes configuration was provided
var ErrNilNodesConfig = errors.New("nil nodes config")

// ErrNilNodesCoordinator signals that a nil nodes coordinator was provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilOutputAntiFloodHandler signals that a nil output antiflood handler was provided
var ErrNilOutputAntiFloodHandler = errors.New("nil output antiflood handler")

// ErrNilPath signals that a nil path was provided
var ErrNilPath = errors.New("nil path provided")

// ErrNilPathHandler signals that a nil path handler was provided
var ErrNilPathHandler = errors.New("nil path handler")

// ErrNilPeerAccounts signals that a nil peer accounts instance was provided
var ErrNilPeerAccounts = errors.New("nil peer accounts")

// ErrNilPeerBlackListHandler signals that a nil peer black list handler was provided
var ErrNilPeerBlackListHandler = errors.New("nil peer black list handler")

// ErrNilPeerHonestyHandler signals that a nil peer honesty handler was provided
var ErrNilPeerHonestyHandler = errors.New("nil peer honesty handler")

// ErrNilPeerShardMapper signals that a nil peer shard mapper was provided
var ErrNilPeerShardMapper = errors.New("nil peer shard mapper")

// ErrNilPeerSignHandler signals that a nil peer sign handler was provided
var ErrNilPeerSignHandler = errors.New("nil peer signature handler")

// ErrNilPendingMiniBlocksHandler signals that a nil pending miniBlocks handler
var ErrNilPendingMiniBlocksHandler = errors.New("nil pending miniBlocks handler")

// ErrNilPoolsHolder signals that a nil pools holder was provided
var ErrNilPoolsHolder = errors.New("nil pools holder")

// ErrNilPrivateKey signals that a nil private key was provided
var ErrNilPrivateKey = errors.New("nil private key")

// ErrNilP2pPrivateKey signals that a nil p2p private key was provided
var ErrNilP2pPrivateKey = errors.New("nil p2p private key")

// ErrNilProcessComponents signals that a nil process components instance was provided
var ErrNilProcessComponents = errors.New("nil process components")

// ErrNilProcessComponentsHolder signals that a nil procss components holder was provided
var ErrNilProcessComponentsHolder = errors.New("nil process components holder")

// ErrNilPubKeyConverter signals that a nil public key converter was provided
var ErrNilPubKeyConverter = errors.New("nil public key converter")

// ErrNilPublicKey signals that a nil public key was provided
var ErrNilPublicKey = errors.New("nil public key")

// ErrNilP2pPublicKey signals that a nil p2p public key was provided
var ErrNilP2pPublicKey = errors.New("nil p2p public key")

// ErrNilRater signals that a nil rater was provided
var ErrNilRater = errors.New("nil rater")

// ErrNilRatingsInfoHandler signals that nil ratings data information was provided
var ErrNilRatingsInfoHandler = errors.New("nil ratings info handler")

// ErrNilRequestHandler signals that a nil request handler was provided
var ErrNilRequestHandler = errors.New("nil request handler")

// ErrNilRequestersFinder signals that a nil requesters finder was provided
var ErrNilRequestersFinder = errors.New("nil requesters finder")

// ErrNilResolversContainer signals that a nil resolvers container was provided
var ErrNilResolversContainer = errors.New("nil resolvers container")

// ErrNilRoundNotifier signals that a nil round notifier has been provided
var ErrNilRoundNotifier = errors.New("nil round notifier")

// ErrNilRoundHandler signals that a nil roundHandler was provided
var ErrNilRoundHandler = errors.New("nil roundHandler")

// ErrNilShardCoordinator signals that a nil shard coordinator was provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator provided")

// ErrNilSoftwareVersion signals that a nil software version was provided
var ErrNilSoftwareVersion = errors.New("nil software version")

// ErrNilStateComponents signals that a nil state components was provided
var ErrNilStateComponents = errors.New("nil state components")

// ErrNilRunTypeComponents signals that a nil runType components was provided
var ErrNilRunTypeComponents = errors.New("nil runType components")

// ErrNilStateComponentsHolder signals that a nil state components holder was provided
var ErrNilStateComponentsHolder = errors.New("nil state components holder")

// ErrNilStatusComponents signals that a nil status components instance was provided
var ErrNilStatusComponents = errors.New("nil status components")

// ErrNilStatusComponentsHolder signals that a nil status components holder was provided
var ErrNilStatusComponentsHolder = errors.New("nil status components holder")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler provided")

// ErrNilHardforkTrigger signals that a nil hardfork trigger was provided
var ErrNilHardforkTrigger = errors.New("nil hardfork trigger")

// ErrNilStorageManagers signals that a nil storage managers instance was provided
var ErrNilStorageManagers = errors.New("nil storage managers")

// ErrNilStorageService signals that a nil storage service was provided
var ErrNilStorageService = errors.New("nil storage service")

// ErrNilSyncTimer signals that a nil ntp synchronized timer was provided
var ErrNilSyncTimer = errors.New("nil sync timer provided")

// ErrNilSystemSCConfig signals that a nil system smart contracts cofiguration was provided
var ErrNilSystemSCConfig = errors.New("nil system smart contract configuration")

// ErrNilTriesContainer signals that a nil tries container was provided
var ErrNilTriesContainer = errors.New("nil tries container provided")

// ErrNilTriesStorageManagers signals that nil tries storage managers were provided
var ErrNilTriesStorageManagers = errors.New("nil tries storage managers provided")

// ErrNilTrieStorageManager signals that a nil trie storage manager was provided
var ErrNilTrieStorageManager = errors.New("nil trie storage manager")

// ErrNilTxLogsProcessor signals that a nil transaction logs processor was provided
var ErrNilTxLogsProcessor = errors.New("nil transaction logs processor")

// ErrNilTxSigner signals that a nil transaction signer was provided
var ErrNilTxSigner = errors.New("nil transaction signer")

// ErrNilTxSignKeyGen signals that a nil transaction signer key generator was provided
var ErrNilTxSignKeyGen = errors.New("nil transaction signing key generator")

// ErrNilP2pKeyGen signals that a nil p2p key generator was provided
var ErrNilP2pKeyGen = errors.New("nil p2p key generator")

// ErrNilTxSignMarshalizer signals that a nil transaction sign marshalizer was provided
var ErrNilTxSignMarshalizer = errors.New("nil transaction marshalizer")

// ErrNilUint64ByteSliceConverter signals that a nil byte slice converter was provided
var ErrNilUint64ByteSliceConverter = errors.New("nil byte slice converter")

// ErrNilValidatorPublicKeyConverter signals that a nil validator public key converter was provided
var ErrNilValidatorPublicKeyConverter = errors.New("validator public key converter")

// ErrNilValidatorsProvider signals a nil validators provider
var ErrNilValidatorsProvider = errors.New("nil validator provider")

// ErrNilValidatorsStatistics signals a that nil validators statistics was handler was provided
var ErrNilValidatorsStatistics = errors.New("nil validator statistics")

// ErrNilVmMarshalizer signals that a nil vm marshalizer was provided
var ErrNilVmMarshalizer = errors.New("nil vm marshalizer")

// ErrNilWatchdog signals that a nil watchdog was provided
var ErrNilWatchdog = errors.New("nil watchdog")

// ErrPollingFunctionRegistration signals an error while registering the polling function registration
var ErrPollingFunctionRegistration = errors.New("cannot register handler func for num of connected peers")

// ErrPublicKeyMismatch signals a mismatch between two public keys that should have matched
var ErrPublicKeyMismatch = errors.New("public key mismatch between the computed and the one read from the file")

// ErrStatusPollingInit signals an error while initializing the application status polling
var ErrStatusPollingInit = errors.New("cannot init AppStatusPolling")

// ErrWrongTypeAssertion signals a wrong type assertion
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNewBootstrapDataProvider signals a new bootstrapDataProvider creation has failed
var ErrNewBootstrapDataProvider = errors.New("bootstrapDataProvider creation has failed")

// ErrBootstrapDataComponentsFactoryCreate signals that an error occured on bootstrapDataComponentsFactory create
var ErrBootstrapDataComponentsFactoryCreate = errors.New("bootstrapDataComponentsFactory create() failed")

// ErrConsensusComponentsFactoryCreate signals that an error occured on consensusComponentsFactory create
var ErrConsensusComponentsFactoryCreate = errors.New("consensusComponentsFactory create failed")

// ErrCoreComponentsFactoryCreate signals that an error occured on coreComponentsFactory create
var ErrCoreComponentsFactoryCreate = errors.New("coreComponentsFactory create failed")

// ErrStatusCoreComponentsFactoryCreate signals that an error occured on statusCoreComponentsFactory create
var ErrStatusCoreComponentsFactoryCreate = errors.New("statusCoreComponentsFactory create failed")

// ErrCryptoComponentsFactoryCreate signals that an error occured on cryptoComponentsFactory create
var ErrCryptoComponentsFactoryCreate = errors.New("cryptoComponentsFactory create failed")

// ErrDataComponentsFactoryCreate signals that an error occured on dataComponentsFactory create
var ErrDataComponentsFactoryCreate = errors.New("dataComponentsFactory create failed")

// ErrNetworkComponentsFactoryCreate signals that an error occured on networkComponentsFactory create
var ErrNetworkComponentsFactoryCreate = errors.New("networkComponentsFactory create failed")

// ErrStateComponentsFactoryCreate signals that an error occured on stateComponentsFactory create
var ErrStateComponentsFactoryCreate = errors.New("stateComponentsFactory create failed")

// ErrRunTypeComponentsFactoryCreate signals that an error occured on runTypeComponentsFactory create
var ErrRunTypeComponentsFactoryCreate = errors.New("runTypeComponentsFactory create failed")

// ErrStatusComponentsFactoryCreate signals that an error occured on statusComponentsFactory create
var ErrStatusComponentsFactoryCreate = errors.New("statusComponentsFactory create failed")

// ErrNewEpochStartBootstrap signals a new epochStartBootstrap creation has failed
var ErrNewEpochStartBootstrap = errors.New("epochStartBootstrap creation has failed")

// ErrNewStorageEpochStartBootstrap signals that a new storageEpochStartBootstrap creation has failed
var ErrNewStorageEpochStartBootstrap = errors.New("storageEpochStartBootstrap creation has failed")

// ErrBootstrap signals the bootstrapping process has failed
var ErrBootstrap = errors.New("bootstrap process has failed")

// ErrNilDataPoolsHolder signals that a nil data pools holder was provided
var ErrNilDataPoolsHolder = errors.New("nil data pools holder")

// ErrNilNodeRedundancyHandler signals that a nil node redundancy handler was provided
var ErrNilNodeRedundancyHandler = errors.New("nil node redundancy handler")

// ErrNilCurrentEpochProvider signals that a nil current epoch provider was provided
var ErrNilCurrentEpochProvider = errors.New("nil current epoch provider")

// ErrNilScheduledTxsExecutionHandler signals that a nil scheduled transactions execution handler was provided
var ErrNilScheduledTxsExecutionHandler = errors.New("nil scheduled transactions execution handler")

// ErrNilScheduledProcessor signals that a nil scheduled processor was provided
var ErrNilScheduledProcessor = errors.New("nil scheduled processor")

// ErrNilTxsSender signals that a nil transactions sender has been provided
var ErrNilTxsSender = errors.New("nil transactions sender has been provided")

// ErrNilProcessStatusHandler signals that a nil process status handler was provided
var ErrNilProcessStatusHandler = errors.New("nil process status handler")

// ErrNilESDTDataStorage signals that a nil esdt data storage has been provided
var ErrNilESDTDataStorage = errors.New("nil esdt data storage")

// ErrNilEnableEpochsHandler signals that a nil enable epochs handler was provided
var ErrNilEnableEpochsHandler = errors.New("nil enable epochs handler")

// ErrSignerNotSupported signals that a not supported signer was provided
var ErrSignerNotSupported = errors.New("signer not supported")

// ErrMissingMultiSignerConfig signals that the multisigner config is missing
var ErrMissingMultiSignerConfig = errors.New("multisigner configuration missing")

// ErrMissingMultiSigner signals that there is no multisigner instance available
var ErrMissingMultiSigner = errors.New("multisigner instance missing")

// ErrMissingEpochZeroMultiSignerConfig signals that the multisigner config for epoch zero is missing
var ErrMissingEpochZeroMultiSignerConfig = errors.New("multisigner configuration missing for epoch zero")

// ErrNilMultiSignerContainer signals that the multisigner container is nil
var ErrNilMultiSignerContainer = errors.New("multisigner container is nil")

// ErrNilCacher signals that a nil cacher has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilSingleSigner is raised when a valid singleSigner is expected but nil used
var ErrNilSingleSigner = errors.New("singleSigner is nil")

// ErrPIDMismatch signals that the pid from the message is different from the cached pid associated to a certain pk
var ErrPIDMismatch = errors.New("pid mismatch")

// ErrSignatureMismatch signals that the signature from the message is different from the cached signature associated to a certain pk
var ErrSignatureMismatch = errors.New("signature mismatch")

// ErrInvalidPID signals that given PID is invalid
var ErrInvalidPID = errors.New("invalid PID")

// ErrInvalidSignature signals that the given signature is invalid
var ErrInvalidSignature = errors.New("invalid signature")

// ErrInvalidHeartbeatV2Config signals that an invalid heartbeat v2 configuration has been provided
var ErrInvalidHeartbeatV2Config = errors.New("invalid heartbeat v2 configuration")

// ErrNilNetworkStatistics signals that a nil network statistics was provided
var ErrNilNetworkStatistics = errors.New("nil network statistics")

// ErrNilResourceMonitor signals that a nil resource monitor was provided
var ErrNilResourceMonitor = errors.New("nil resource monitor")

// ErrNilTrieSyncStatistics signals that a nil trie sync statistics was provided
var ErrNilTrieSyncStatistics = errors.New("nil trie sync statistics")

// ErrNilAppStatusHandler signals that a nil app status handler was provided
var ErrNilAppStatusHandler = errors.New("nil app status handler")

// ErrNilStatusMetrics signals that a nil status metrics was provided
var ErrNilStatusMetrics = errors.New("nil status metrics")

// ErrNilPersistentHandler signals that a nil persistent handler was provided
var ErrNilPersistentHandler = errors.New("nil persistent handler")

// ErrNilGenesisNodesSetupHandler signals that a nil genesis nodes setup handler has been provided
var ErrNilGenesisNodesSetupHandler = errors.New("nil genesis nodes setup handler")

// ErrNilManagedPeersHolder signals that a nil managed peers holder has been provided
var ErrNilManagedPeersHolder = errors.New("nil managed peers holder")

// ErrNilManagedPeersMonitor signals that a nil managed peers monitor has been provided
var ErrNilManagedPeersMonitor = errors.New("nil managed peers monitor")

// ErrUnimplementedConsensusModel signals an unimplemented consensus model
var ErrUnimplementedConsensusModel = errors.New("unimplemented consensus model")

// ErrNilPeersRatingHandler signals that a nil peers rating handler implementation has been provided
var ErrNilPeersRatingHandler = errors.New("nil peers rating handler")

// ErrNilPeersRatingMonitor signals that a nil peers rating monitor implementation has been provided
var ErrNilPeersRatingMonitor = errors.New("nil peers rating monitor")

// ErrNilLogger signals that a nil logger instance has been provided
var ErrNilLogger = errors.New("nil logger")

// ErrNilShuffleOutCloser signals that a nil shuffle out closer has been provided
var ErrNilShuffleOutCloser = errors.New("nil shuffle out closer")

// ErrNilHistoryRepository signals that history processor is nil
var ErrNilHistoryRepository = errors.New("history repository is nil")

// ErrNilMissingTrieNodesNotifier signals that a nil missing trie nodes notifier was provided
var ErrNilMissingTrieNodesNotifier = errors.New("nil missing trie nodes notifier")

// ErrInvalidTrieNodeVersion signals that an invalid trie node version has been provided
var ErrInvalidTrieNodeVersion = errors.New("invalid trie node version")

// ErrNilTrieMigrator signals that a nil trie migrator has been provided
var ErrNilTrieMigrator = errors.New("nil trie migrator")

// ErrNilAddress defines the error when trying to work with a nil address
var ErrNilAddress = errors.New("nil address")

// ErrInsufficientFunds signals the funds are insufficient for the move balance operation but the
// transaction fee is covered by the current balance
var ErrInsufficientFunds = errors.New("insufficient funds")

// ErrOperationNotPermitted signals that operation is not permitted
var ErrOperationNotPermitted = errors.New("operation in account not permitted")

// ErrInvalidAddressLength signals that address length is invalid
var ErrInvalidAddressLength = errors.New("invalid address length")

// ErrNilTrackableDataTrie signals that a nil trackable data trie has been provided
var ErrNilTrackableDataTrie = errors.New("nil trackable data trie")

// ErrNilTrieLeafParser signals that a nil trie leaf parser has been provided
var ErrNilTrieLeafParser = errors.New("nil trie leaf parser")

// ErrNilTrie signals that a trie is nil and no operation can be made
var ErrNilTrie = errors.New("trie is nil")

// ErrNilBLSPublicKey signals that the provided BLS public key is nil
var ErrNilBLSPublicKey = errors.New("bls public key is nil")

// ErrEmptyAddress defines the error when trying to work with an empty address
var ErrEmptyAddress = errors.New("empty Address")

// ErrInvalidNodeOperationMode signals that an invalid node operation mode has been provided
var ErrInvalidNodeOperationMode = errors.New("invalid node operation mode")

// ErrNilSentSignatureTracker defines the error for setting a nil SentSignatureTracker
var ErrNilSentSignatureTracker = errors.New("nil sent signature tracker")

// ErrNilNodesCoordinatorFactory signals that a nil nodes coordinator factory has been provided
var ErrNilNodesCoordinatorFactory = errors.New("nil nodes coordinator factory provided")

// ErrNilShardCoordinatorFactory signals that a nil shard coordinator factory has been provided
var ErrNilShardCoordinatorFactory = errors.New("nil shard coordinator factory provided")

// ErrNilGenesisBlockFactory signals that a nil genesis block factory has been provided
var ErrNilGenesisBlockFactory = errors.New("nil genesis block factory has been provided")

// ErrNilNodesSetupFactory signals that a nil nodes setup factory has been provided
var ErrNilNodesSetupFactory = errors.New("nil nodes setup factory provided")

// ErrNilRatingsDataFactory signals that a nil ratings data factory has been provided
var ErrNilRatingsDataFactory = errors.New("nil ratings data factory provided")

// ErrNilGenesisMetaBlockChecker signals that a nil genesis meta block checker has been provided
var ErrNilGenesisMetaBlockChecker = errors.New("nil genesis meta block checker has been provided")

// ErrGenesisMetaBlockDoesNotExist signals that genesis meta block does not exist
var ErrGenesisMetaBlockDoesNotExist = errors.New("genesis meta block does not exist")

// ErrInvalidGenesisMetaBlock signals that genesis meta block should be of type meta header handler
var ErrInvalidGenesisMetaBlock = errors.New("genesis meta block invalid, should be of type meta header handler")

// ErrGenesisMetaBlockOnSovereign signals that genesis meta block was found on sovereign chain
var ErrGenesisMetaBlockOnSovereign = errors.New("genesis meta block was found on sovereign chain")

// ErrNilShardRequesterContainerFactory signals that a nil shard requester container factory has been provided
var ErrNilShardRequesterContainerFactory = errors.New("nil shard shard requester container factory provided")

// ErrNilRequesterContainerFactoryCreator signals that a nil requester container factory creator has been provided
var ErrNilRequesterContainerFactoryCreator = errors.New("nil requester container factory creator provided")

// ErrNilInterceptorsContainerFactoryCreator signals that a nil interceptors container factory creator has been provided
var ErrNilInterceptorsContainerFactoryCreator = errors.New("nil interceptors container factory creator has been provided")

// ErrNilShardInterceptorsContainerFactory signals that a nil shard interceptors container factory has been provided
var ErrNilShardInterceptorsContainerFactory = errors.New("nil shard interceptors container factory has been provided")

// ErrNilIncomingHeaderSubscriber signals that a nil incoming header subscriber has been provided
var ErrNilIncomingHeaderSubscriber = errors.New("nil incoming header subscriber has been provided")

// ErrNilShardResolversContainerFactory signals that a nil shard resolvers container factory has been provided
var ErrNilShardResolversContainerFactory = errors.New("nil shard resolvers container factory has been provided")

// ErrNilShardResolversContainerFactoryCreator signals that a nil shard resolvers container factory creator has been provided
var ErrNilShardResolversContainerFactoryCreator = errors.New("nil shard resolvers container factory creator has been provided")

// ErrInvalidReceivedExtendedShardHeader signals that an invalid extended shard header has been intercepted when requested
var ErrInvalidReceivedExtendedShardHeader = errors.New("invalid extended shard header has been intercepted")

// ErrNilTxPreProcessorCreator signals that a nil tx pre-processor creator has been provided
var ErrNilTxPreProcessorCreator = errors.New("nil tx pre-processor creator has been provided")

// ErrNilNode signals that a nil node was provided
var ErrNilNode = errors.New("nil node")

// ErrNilOutgoingOperationsFormatter signals that a nil outgoing operations formatter has been provided
var ErrNilOutgoingOperationsFormatter = errors.New("nil outgoing operations formatter has been provided")

// ErrNilOutGoingOperationsPool signals that a nil outgoing operations pool has been provided
var ErrNilOutGoingOperationsPool = errors.New("nil outgoing operations pool has been provided")

// ErrNilDataCodec signals that a nil outgoing operations pool has been provided
var ErrNilDataCodec = errors.New("nil data codec has been provided")

// ErrNilExtraSubRoundSigner signals that a nil extra subround signer has been provided
var ErrNilExtraSubRoundSigner = errors.New("nil extra subround signer has been provided")

// ErrNilStartRoundExtraSignersHolder signals that a nil start round extra signers holder has been provided
var ErrNilStartRoundExtraSignersHolder = errors.New("nil start round extra signers holder has been provided")

// ErrNilSignatureRoundExtraSignersHolder signals that a nil signature round extra signers holder has been provided
var ErrNilSignatureRoundExtraSignersHolder = errors.New("nil signature round extra signers holder has been provided")

// ErrNilEndRoundExtraSignersHolder signals that a nil end round extra signers holder has been provided
var ErrNilEndRoundExtraSignersHolder = errors.New("nil end round extra signers holder has been provided")

// ErrExtraSignerIdAlreadyExists signals that an extra signer with the same id has been already registered
var ErrExtraSignerIdAlreadyExists = errors.New("an extra signer with the same id has been already registered")

// ErrNilExtraHeaderSigVerifierHolder signals that a nil extra sig verifier holder has been provided
var ErrNilExtraHeaderSigVerifierHolder = errors.New("nil extra sig verifier holder has been provided")

// ErrNilConsensusMessage signals that a nil consensus message has been provided
var ErrNilConsensusMessage = errors.New("nil consensus message has been provided")

// ErrNilExtraSignersHolder signals that a nil extra signers holder has been provided
var ErrNilExtraSignersHolder = errors.New("nil extra signer holder has been provided")

// ErrNilSubRoundEndV2Creator signals that a nil sub round end v2 creator has been provided
var ErrNilSubRoundEndV2Creator = errors.New("nil sub round end v2 creator has been provided")

// ErrNilBridgeOpHandler signals that a nil bridge operation handler has been provided
var ErrNilBridgeOpHandler = errors.New("nil bridge operation handler has been provided")

// ErrOutGoingOperationsNotFound signals that an outgoing operation could not be found
var ErrOutGoingOperationsNotFound = errors.New("outgoing operation could not be found")

// ErrInvalidTypeConversion signals that a type conversion has failed
var ErrInvalidTypeConversion = errors.New("invalid type conversion")

// ErrNilShardBootstrapFactory signals that a nil shard bootstrap factory has been provided
var ErrNilShardBootstrapFactory = errors.New("nil shard bootstrap factory")

// ErrNilShardStorageBootstrapperFactory signals that a nil shard storage bootstrapper factory has been provided
var ErrNilShardStorageBootstrapperFactory = errors.New("nil shard storage bootstrapper factory")

// ErrNilResolverRequestFactoryHandler signals that a nil resolver request factory handler has been provided
var ErrNilResolverRequestFactoryHandler = errors.New("nil resolver request factory handler")

// ErrNilBlockChainHookFactory signals that a nil blockchain hook factory has been provided
var ErrNilBlockChainHookFactory = errors.New("nil blockchain hook factory")

// ErrNilEpochStartBootstrapperFactory signals that a nil epoch start bootstrapper factory has been provided
var ErrNilEpochStartBootstrapperFactory = errors.New("nil epoch start bootstrapper factory")

// ErrNilBlockProcessorFactory signals that a nil block processor factory has been provided
var ErrNilBlockProcessorFactory = errors.New("nil block processor factory")

// ErrNilBlockChainHookHandlerCreator signals that a nil blockchain hook handler creator has been provided
var ErrNilBlockChainHookHandlerCreator = errors.New("nil blockchain hook handler creator")

// ErrNilEpochStartBootstrapperCreator signals that a nil epoch start bootstrapper creator has been provided
var ErrNilEpochStartBootstrapperCreator = errors.New("nil epoch start bootstrapper creator")

// ErrNilBootstrapperFromStorageCreator signals that a nil bootstrapper from storage creator has been provided
var ErrNilBootstrapperFromStorageCreator = errors.New("nil bootstrapper from storage creator")

// ErrNilBootstrapperCreator signals that a nil bootstrapper creator has been provided
var ErrNilBootstrapperCreator = errors.New("nil bootstrapper creator")

// ErrNilBlockProcessorCreator signals that a nil block processor creator has been provided
var ErrNilBlockProcessorCreator = errors.New("nil block processor creator")

// ErrNilForkDetectorCreator signals that a nil fork detector creator has been provided
var ErrNilForkDetectorCreator = errors.New("nil fork detector creator")

// ErrNilBlockTrackerCreator signals that a nil block tracker creator has been provided
var ErrNilBlockTrackerCreator = errors.New("nil block tracker creator")

// ErrNilRequestHandlerCreator signals that a nil request handler creator has been provided
var ErrNilRequestHandlerCreator = errors.New("nil request handler creator")

// ErrNilHeaderValidatorCreator signals that a nil header validator creator has been provided
var ErrNilHeaderValidatorCreator = errors.New("nil header validator creator")

// ErrNilScheduledTxsExecutionCreator signals that a nil scheduled txs execution creator has been provided
var ErrNilScheduledTxsExecutionCreator = errors.New("nil scheduled txs execution creator")

// ErrNilTransactionCoordinatorCreator signals that a nil transaction coordinator creator has been provided
var ErrNilTransactionCoordinatorCreator = errors.New("nil transaction coordinator creator")

// ErrNilValidatorStatisticsProcessorCreator signals that a nil validator statistics processor creator has been provided
var ErrNilValidatorStatisticsProcessorCreator = errors.New("nil validator statistics processor creator")

// ErrNilAdditionalStorageServiceCreator signals that a nil additional storage service creator has been provided
var ErrNilAdditionalStorageServiceCreator = errors.New("nil additional storage service creator")

// ErrNilSCProcessorCreator signals that a nil sc processor creator has been provided
var ErrNilSCProcessorCreator = errors.New("nil sc processor creator")

// ErrNilSCResultsPreProcessorCreator signals that a nil sc results pre processor creator has been provided
var ErrNilSCResultsPreProcessorCreator = errors.New("nil sc results pre processor creator")

// ErrNilFunction signals that a nil function has been provided
var ErrNilFunction = errors.New("nil function")

// ErrEmptyBaseToken signals that an empty base token has been provided
var ErrEmptyBaseToken = errors.New("empty base token")

// ErrNilESDTAsBalanceHandler signals that a nil esdt as balance handler was provided
var ErrNilESDTAsBalanceHandler = errors.New("nil esdt as balance handler provided")

// ErrNilSerializer signals that a nil codec serializer has been provided
var ErrNilSerializer = errors.New("nil codec serializer")

// ErrNilTopicsChecker signals that a nil topics checker has been provided
var ErrNilTopicsChecker = errors.New("nil topics checker")

// ErrNilOperationsHasher signals that a nil outgoing operations hasher has been provided
var ErrNilOperationsHasher = errors.New("nil outgoing operations hasher")

// ErrNilEpochSystemSCProcessor defines the error for setting a nil EpochSystemSCProcessor
var ErrNilEpochSystemSCProcessor = errors.New("nil epoch system SC processor")

// ErrNilBlockChainHookCreator signals that a nil blockchain hook creator has been provided
var ErrNilBlockChainHookCreator = errors.New("nil blockchain hook creator has been provided")

// ErrNilVMContextCreator signals that a nil vm context creator has been provided
var ErrNilVMContextCreator = errors.New("nil vm context creator has been provided")

// ErrInvalidConsensusModel signals that an invalid consensus model has been provided
var ErrInvalidConsensusModel = errors.New("invalid consensus model")

// ErrNilVmContainerMetaFactoryCreator signals that a nil vm container meta factory creator has been provided
var ErrNilVmContainerMetaFactoryCreator = errors.New("nil vm container meta factory creator")

// ErrNilVmContainerShardFactoryCreator signals that a nil vm container shard factory creator has been provided
var ErrNilVmContainerShardFactoryCreator = errors.New("nil vm container shard factory creator")

// ErrNilAccountsCreator signals that a nil accounts creator has been provided
var ErrNilAccountsCreator = errors.New("nil accounts creator")

// ErrNilInitialAccounts signals that a nil initial accounts has been provided
var ErrNilInitialAccounts = errors.New("nil initial accounts")
