package factory

import "errors"

// ErrNilEconomicsData signals that a nil economics data handler has been provided
var ErrNilEconomicsData = errors.New("nil economics data provided")

// ErrNilCoreComponents signals that nil core components have been provided
var ErrNilCoreComponents = errors.New("nil core components provided")

// ErrNilCryptoComponents signals that a nil crypto components has been provided
var ErrNilCryptoComponents = errors.New("nil crypto components provided")

// ErrNilDataComponents signals that a nil data components has been provided
var ErrNilDataComponents = errors.New("nil data components provided")

// ErrNilTriesContainer signals that a nil tries container has been provided
var ErrNilTriesContainer = errors.New("nil tries container provided")

// ErrNilTriesStorageManagers signals that nil tries storage managers have been provided
var ErrNilTriesStorageManagers = errors.New("nil tries storage managers providedd")

// ErrNilShardCoordinator signals that nil core components have been provided
var ErrNilShardCoordinator = errors.New("nil shard coordinator provided")

// ErrNilPathManager signals that a nil path manager has been provided
var ErrNilPathManager = errors.New("nil path manager provided")

// ErrNilPath signals that a nil/empty path was provided
var ErrNilPath = errors.New("nil path provided")

// ErrNilKeyLoader signals that a nil key loader has been provided
var ErrNilKeyLoader = errors.New("nil key loader")

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("nil marshalizer provided")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("nil hasher provided")

// ErrNilEpochStartNotifier signals that a nil epoch start notifier has been provided
var ErrNilEpochStartNotifier = errors.New("nil epoch start notifier provided")

// ErrHasherCreation signals that the hasher cannot be created based on provided data
var ErrHasherCreation = errors.New("error creating hasher")

// ErrMarshalizerCreation signals that the marshalizer cannot be created based on provided data
var ErrMarshalizerCreation = errors.New("error creating marshalizer")

// ErrAccountsAdapterCreation signals that the accounts adapter cannot be created based on provided data
var ErrAccountsAdapterCreation = errors.New("error creating accounts adapter")

// ErrPublicKeyMismatch signals that the read public key mismatch the one read
var ErrPublicKeyMismatch = errors.New("public key mismatch between the computed and the one read from the file")

// ErrBlockchainCreation signals that the blockchain cannot be created
var ErrBlockchainCreation = errors.New("can not create blockchain")

// ErrDataStoreCreation signals that the data store cannot be created
var ErrDataStoreCreation = errors.New("can not create data store")

// ErrDataPoolCreation signals that the data pool cannot be created
var ErrDataPoolCreation = errors.New("can not create data pool")

// ErrInvalidConsensusConfig signals that an invalid consensus type is specified in the configuration file
var ErrInvalidConsensusConfig = errors.New("invalid consensus type provided in config file")

// ErrMultiSigHasherMissmatch signals that an invalid multisig hasher was provided
var ErrMultiSigHasherMissmatch = errors.New("wrong multisig hasher provided for bls consensus type")

// ErrMissingMultiHasherConfig signals that the multihasher type isn't specified in the configuration file
var ErrMissingMultiHasherConfig = errors.New("no multisig hasher provided in config file")

// ErrNilStatusHandler signals that a nil status handler has been provided
var ErrNilStatusHandler = errors.New("nil status handler provided")

// ErrNilSyncTimer signals that a nil sync timer has been provided
var ErrNilSyncTimer = errors.New("nil sync timer provided")

// ErrWrongTypeAssertion signals that a wrong type assertion occurred
var ErrWrongTypeAssertion = errors.New("wrong type assertion")

// ErrNilAccountsParser signals that a nil accounts parser has been provided
var ErrNilAccountsParser = errors.New("nil accounts parser")

// ErrNilSmartContractParser signals that a nil smart contract parser has been provided
var ErrNilSmartContractParser = errors.New("nil smart contract parser")

// ErrNilNodesConfig signals that a nil nodes config has been provided
var ErrNilNodesConfig = errors.New("nil nodes config")

// ErrNilGasSchedule signals that a nil gas schedule has been provided
var ErrNilGasSchedule = errors.New("nil gas schedule")

// ErrNilRounder signals that a nil rounder has been provided
var ErrNilRounder = errors.New("nil rounder")

// ErrNilNodesCoordinator signals that nil nodes coordinator has been provided
var ErrNilNodesCoordinator = errors.New("nil nodes coordinator")

// ErrNilBootstrapComponentsHolder signals that a nil bootstrap components holder has been provided
var ErrNilBootstrapComponentsHolder = errors.New("nil bootstrap components holder")

// ErrNilConsensusComponentsHolder signals that a nil consensus components holder has been provided
var ErrNilConsensusComponentsHolder = errors.New("nil consensus components holder")

// ErrNilCoreComponentsHolder signals that a nil core components holder has been provided
var ErrNilCoreComponentsHolder = errors.New("nil core components holder")

// ErrNilCryptoComponentsHolder signals that a nil crypto components holder has been provided
var ErrNilCryptoComponentsHolder = errors.New("nil crypto components holder")

// ErrNilDataComponentsHolder signals that a nil data components holder has been provided
var ErrNilDataComponentsHolder = errors.New("nil data components holder")

// ErrNilHeartbeatComponentsHolder signals that a nil heartbeat components holder has been provided
var ErrNilHeartbeatComponentsHolder = errors.New("nil heartbeat components holder")

// ErrNilNetworkComponentsHolder signals that a nil network components holder has been provided
var ErrNilNetworkComponentsHolder = errors.New("nil network components holder")

// ErrNilProcessComponentsHolder signals that a nil process components holder was provided
var ErrNilProcessComponentsHolder = errors.New("nil process components holder")

// ErrNilStateComponentsHolder signals that a nil state components holder has been provided
var ErrNilStateComponentsHolder = errors.New("nil state components holder")

// ErrNilStatusComponentsHolder signals that a nil status components holder has been provided
var ErrNilStatusComponentsHolder = errors.New("nil status components holder")

// ErrNilMessenger signals a nil messenger was provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilCoreServiceContainer signals that a nil core service container has been provided
var ErrNilCoreServiceContainer = errors.New("nil core service container")

// ErrNilRequestedItemHandler signals that a nil requested item handler has been provided
var ErrNilRequestedItemHandler = errors.New("nil requested item handler")

// ErrNilWhiteListHandler signals that a nil white list handler has been provided
var ErrNilWhiteListHandler = errors.New("nil white list handler")

// ErrNilWhiteListVerifiedTxs signals that a nil white list verifies txs has been provided
var ErrNilWhiteListVerifiedTxs = errors.New("nil white list verified txs")

// ErrNilEpochStartConfig signals that a nil epoch start configuration has been provided
var ErrNilEpochStartConfig = errors.New("nil epoch start configuration")

// ErrNilRater signals that a nil rater has been provided
var ErrNilRater = errors.New("nil rater")

// ErrNilRatingData signals that a nil rating data has been provided
var ErrNilRatingData = errors.New("nil rating data")

// ErrNilPubKeyConverter signals that a nil public key converter has been provided
var ErrNilPubKeyConverter = errors.New("nil public key converter")

// ErrNilSystemSCConfig signals that a nil system smart contract configuration has been provided
var ErrNilSystemSCConfig = errors.New("nil system smart contract configuration")

// ErrInvalidRoundDuration signals that an invalid round duration has been provided
var ErrInvalidRoundDuration = errors.New("invalid round duration provided")

// ErrNilElasticOptions signals that nil elastic options have been provided
var ErrNilElasticOptions = errors.New("nil elastic options")

// ErrNilStatusHandlersUtils signals that a nil status handlers utils has been provided
var ErrNilStatusHandlersUtils = errors.New("nil status handlers utils")

// ErrValidatorAlreadySet signals that a topic validator has already been set
var ErrValidatorAlreadySet = errors.New("topic validator has already been set")

// ErrNilGenesisNodesSetup signals that a nil genesis nodes setup
var ErrNilGenesisNodesSetup = errors.New("nil genesis nodes setup")

// ErrInvalidWorkingDir signals an invalid working directory
var ErrInvalidWorkingDir = errors.New("invalid working directory")

// ErrNilShuffler signals a nil nodes shuffler
var ErrNilShuffler = errors.New("nil nodes shuffler")

// ErrGenesisBlockNotInitialized signals that genesis block is not initialized
var ErrGenesisBlockNotInitialized = errors.New("genesis block is not initialized")
