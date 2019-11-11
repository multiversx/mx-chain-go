package dataRetriever

import (
	"errors"
)

// ErrNilMessage signals that a nil message has been received
var ErrNilMessage = errors.New("nil message")

// ErrNoConnectedPeerToSendRequest signals that the connected peers list is empty and can not send request
var ErrNoConnectedPeerToSendRequest = errors.New("connected peers list is empty. Can not send request")

// ErrNilAccountsAdapter defines the error when trying to use a nil AccountsAddapter
var ErrNilAccountsAdapter = errors.New("nil AccountsAdapter")

// ErrNilHasher signals that an operation has been attempted to or with a nil hasher implementation
var ErrNilHasher = errors.New("nil Hasher")

// ErrNilAddressConverter signals that an operation has been attempted to or with a nil AddressConverter implementation
var ErrNilAddressConverter = errors.New("nil AddressConverter")

// ErrNilAddressContainer signals that an operation has been attempted to or with a nil AddressContainer implementation
var ErrNilAddressContainer = errors.New("nil AddressContainer")

// ErrNilValue signals the value is nil
var ErrNilValue = errors.New("nil value")

// ErrNilBlockChain signals that an operation has been attempted to or with a nil blockchain
var ErrNilBlockChain = errors.New("nil block chain")

// ErrNilTxBlockBody signals that an operation has been attempted to or with a nil block body
var ErrNilTxBlockBody = errors.New("nil block body")

// ErrNilBlockHeader signals that an operation has been attempted to or with a nil block header
var ErrNilBlockHeader = errors.New("nil block header")

// ErrNilPublicKey signals that a operation has been attempted with a nil public key
var ErrNilPublicKey = errors.New("nil public key")

// ErrNilSignature signals that a operation has been attempted with a nil signature
var ErrNilSignature = errors.New("nil signature")

// ErrEmptyMiniBlockSlice signals that an operation has been attempted with an empty mini block slice
var ErrEmptyMiniBlockSlice = errors.New("empty mini block slice")

// ErrInvalidShardId signals that the shard id is invalid
var ErrInvalidShardId = errors.New("invalid shard id")

// ErrNilMarshalizer signals that an operation has been attempted to or with a nil Marshalizer implementation
var ErrNilMarshalizer = errors.New("nil Marshalizer")

// ErrNilStore signals that the provided storage service is nil
var ErrNilStore = errors.New("nil data storage service")

// ErrNilRounder signals that an operation has been attempted to or with a nil Rounder implementation
var ErrNilRounder = errors.New("nil Rounder")

// ErrNilMessenger signals that a nil Messenger object was provided
var ErrNilMessenger = errors.New("nil Messenger")

// ErrNilTxDataPool signals that a nil transaction pool has been provided
var ErrNilTxDataPool = errors.New("nil transaction data pool")

// ErrNilUnsignedTransactionPool signals that a nil unsigned transactions pool has been provided
var ErrNilUnsignedTransactionPool = errors.New("nil unsigned transactions data pool")

// ErrNilRewardTransactionPool signals that a nil reward transactions pool has been provided
var ErrNilRewardTransactionPool = errors.New("nil reward transaction data pool")

// ErrNilHeadersDataPool signals that a nil header pool has been provided
var ErrNilHeadersDataPool = errors.New("nil headers data pool")

// ErrNilHeadersNoncesDataPool signals that a nil header - nonce cache
var ErrNilHeadersNoncesDataPool = errors.New("nil headers nonces cache")

// ErrNilShardCoordinator signals that an operation has been attempted to or with a nil shard coordinator
var ErrNilShardCoordinator = errors.New("nil shard coordinator")

// ErrNilSingleSigner signals that a nil single signer is used
var ErrNilSingleSigner = errors.New("nil single signer")

// ErrNilDataToProcess signals that nil data was provided
var ErrNilDataToProcess = errors.New("nil data to process")

// ErrNilTxStorage signals that a nil transaction storage has been provided
var ErrNilTxStorage = errors.New("nil transaction storage")

// ErrNilHeadersStorage signals that a nil header storage has been provided
var ErrNilHeadersStorage = errors.New("nil headers storage")

// ErrNilHeadersNoncesStorage signals that a nil header-nonce storage has been provided
var ErrNilHeadersNoncesStorage = errors.New("nil headers nonces storage")

// ErrNilResolverSender signals that a nil resolver sender object has been provided
var ErrNilResolverSender = errors.New("nil resolver sender")

// ErrInvalidNonceByteSlice signals that an invalid byte slice has been provided
// and an uint64 can not be decoded from that byte slice
var ErrInvalidNonceByteSlice = errors.New("invalid nonce byte slice")

// ErrResolveTypeUnknown signals that an unknown resolve type was provided
var ErrResolveTypeUnknown = errors.New("unknown resolve type")

// ErrNilBlockBodyPool signals that a nil block body pool has been provided
var ErrNilBlockBodyPool = errors.New("nil block body pool")

// ErrNilBlockBodyStorage signals that a nil block body storage has been provided
var ErrNilBlockBodyStorage = errors.New("nil block body storage")

// ErrNilDataPoolHolder signals that the data pool holder is nil
var ErrNilDataPoolHolder = errors.New("nil data pool holder")

// ErrNilContainerElement signals when trying to add a nil element in the container
var ErrNilContainerElement = errors.New("element cannot be nil")

// ErrInvalidContainerKey signals that an element does not exist in the container's map
var ErrInvalidContainerKey = errors.New("element does not exist in container")

// ErrContainerKeyAlreadyExists signals that an element was already set in the container's map
var ErrContainerKeyAlreadyExists = errors.New("provided key already exists in container")

// ErrNilUint64ByteSliceConverter signals that a nil byte slice converter was provided
var ErrNilUint64ByteSliceConverter = errors.New("nil byte slice converter")

// ErrNilResolverContainer signals that a nil resolver container was provided
var ErrNilResolverContainer = errors.New("nil resolver container")

// ErrUnmarshalMBHashes signals the value is nil
var ErrUnmarshalMBHashes = errors.New("could not unmarshal miniblock hashes")

// ErrInvalidRequestType signals that a request on a topic sends an invalid type
var ErrInvalidRequestType = errors.New("invalid request type")

// ErrWrongTypeInContainer signals that a wrong type of object was found in container
var ErrWrongTypeInContainer = errors.New("wrong type of object inside container")

// ErrLenMismatch signals that 2 or more slices have different lengths
var ErrLenMismatch = errors.New("lengths mismatch")

// ErrNilPeerChangeBlockDataPool signals that a nil peer change pool has been provided
var ErrNilPeerChangeBlockDataPool = errors.New("nil peer change block data pool")

// ErrNilTxBlockDataPool signals that a nil tx block body pool has been provided
var ErrNilTxBlockDataPool = errors.New("nil tx block data pool")

// ErrNilCacher signals that a nil cache has been provided
var ErrNilCacher = errors.New("nil cacher")

// ErrNilMetaBlockPool signals that a nil meta block data pool was provided
var ErrNilMetaBlockPool = errors.New("nil meta block data pool")

// ErrNilMiniBlockHashesPool signals that a nil meta block data pool was provided
var ErrNilMiniBlockHashesPool = errors.New("nil meta block mini block hashes data pool")

// ErrNilShardHeaderPool signals that a nil meta block data pool was provided
var ErrNilShardHeaderPool = errors.New("nil meta block shard header data pool")

// ErrNilMetaBlockNoncesPool signals that a nil meta block data pool was provided
var ErrNilMetaBlockNoncesPool = errors.New("nil meta block nonces data pool")

// ErrNoSuchStorageUnit defines the error for using an invalid storage unit
var ErrNoSuchStorageUnit = errors.New("no such unit type")

// ErrNilRandomizer signals that a nil randomizer has been provided
var ErrNilRandomizer = errors.New("nil randomizer")

// ErrRequestTypeNotImplemented signals that a not implemented type of request has been received
var ErrRequestTypeNotImplemented = errors.New("request type is not implemented")

// ErrNilDataPacker signals that a nil data packer has been provided
var ErrNilDataPacker = errors.New("nil data packer provided")

// ErrNilResolverFinder signals that a nil resolver finder has been provided
var ErrNilResolverFinder = errors.New("nil resolvers finder")

// ErrEmptyTxRequestTopic signals that an empty transaction topic has been provided
var ErrEmptyTxRequestTopic = errors.New("empty transaction request topic")

// ErrEmptyScrRequestTopic signals that an empty smart contract result topic has been provided
var ErrEmptyScrRequestTopic = errors.New("empty smart contract result request topic")

// ErrEmptyRewardTxRequestTopic signals that an empty reward transaction topic has been provided
var ErrEmptyRewardTxRequestTopic = errors.New("empty rewards transactions request topic")

// ErrEmptyMiniBlockRequestTopic signals that an empty miniblock topic has been provided
var ErrEmptyMiniBlockRequestTopic = errors.New("empty miniblock request topic")

// ErrEmptyShardHeaderRequestTopic signals that an empty shard header topic has been provided
var ErrEmptyShardHeaderRequestTopic = errors.New("empty shard header request topic")

// ErrEmptyMetaHeaderRequestTopic signals that an empty meta header topic has been provided
var ErrEmptyMetaHeaderRequestTopic = errors.New("empty meta header request topic")

// ErrInvalidMaxTxRequest signals that max tx request is too small
var ErrInvalidMaxTxRequest = errors.New("max tx request number is invalid")

// ErrNilPeerListCreator signals that a nil peer list creator implementation has been provided
var ErrNilPeerListCreator = errors.New("nil peer list creator provided")

// ErrNilCurrBlockTxs signals that nil current blocks txs holder was provided
var ErrNilCurrBlockTxs = errors.New("nil current block txs holder")
