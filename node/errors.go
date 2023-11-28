package node

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("trying to set nil marshalizer")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilPubkeyConverter signals that a nil public key converter has been provided
var ErrNilPubkeyConverter = errors.New("trying to use a nil pubkey converter")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("trying to set nil private key")

// ErrZeroRoundDurationNotSupported signals that 0 seconds round duration is not supported
var ErrZeroRoundDurationNotSupported = errors.New("0 round duration time is not supported")

// ErrNegativeOrZeroConsensusGroupSize signals that 0 elements consensus group is not supported
var ErrNegativeOrZeroConsensusGroupSize = errors.New("group size should be a strict positive number")

// ErrNilDataPool signals that a nil data pool has been provided
var ErrNilDataPool = errors.New("trying to set nil data pool")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilSingleSig signals that a nil singlesig object has been provided
var ErrNilSingleSig = errors.New("trying to set nil singlesig")

// ErrNilPeerDenialEvaluator signals that a nil peer denial evaluator was provided
var ErrNilPeerDenialEvaluator = errors.New("nil peer denial evaluator")

// ErrNilRequestedItemsHandler signals that a nil requested items handler was provided
var ErrNilRequestedItemsHandler = errors.New("nil requested items handler")

// ErrSystemBusyGeneratingTransactions signals that to many transactions are trying to get generated
var ErrSystemBusyGeneratingTransactions = errors.New("system busy while generating bulk transactions")

// ErrInvalidValue signals that an invalid value has been provided such as NaN to an integer field
var ErrInvalidValue = errors.New("invalid value")

// ErrInvalidSignatureLength signals that an invalid signature length has been provided
var ErrInvalidSignatureLength = errors.New("invalid signature length")

// ErrInvalidAddressLength signals that an invalid address length has been provided
var ErrInvalidAddressLength = errors.New("invalid address length")

// ErrInvalidChainIDInTransaction signals that an invalid chain id has been provided in transaction
var ErrInvalidChainIDInTransaction = errors.New("invalid chain ID")

// ErrInvalidSenderUsernameLength signals that the length of the sender username is invalid
var ErrInvalidSenderUsernameLength = errors.New("invalid sender username length")

// ErrInvalidReceiverUsernameLength signals that the length of the receiver username is invalid
var ErrInvalidReceiverUsernameLength = errors.New("invalid receiver username length")

// ErrDataFieldTooBig signals that the data field is too big
var ErrDataFieldTooBig = errors.New("data field is too big")

// ErrNilNodeStopChannel signals that a nil channel for node process stop has been provided
var ErrNilNodeStopChannel = errors.New("nil node stop channel")

// ErrNilESDTNFTStorageHandler signals that a nil esdt and nft storage handler has been provided
var ErrNilESDTNFTStorageHandler = errors.New("nil esdt and nft storage handler")

// ErrNilQueryHandler signals that a nil query handler has been provided
var ErrNilQueryHandler = errors.New("nil query handler")

// ErrQueryHandlerAlreadyExists signals that the query handler is already registered
var ErrQueryHandlerAlreadyExists = errors.New("query handler already exists")

// ErrEmptyQueryHandlerName signals that an empty string can not be used to be used in the query handler container
var ErrEmptyQueryHandlerName = errors.New("empty query handler name")

// ErrUnknownPeerID signals that the provided peer is unknown by the current node
var ErrUnknownPeerID = errors.New("unknown peer ID")

// ErrInvalidTransactionVersion signals that an invalid transaction version has been provided
var ErrInvalidTransactionVersion = errors.New("invalid transaction version")

// ErrNilBootstrapComponents signals that a nil bootstrap components instance has been provided
var ErrNilBootstrapComponents = errors.New("nil bootstrap componennts")

// ErrNilCoreComponents signals that a nil core components instance has been provided
var ErrNilCoreComponents = errors.New("nil core components")

// ErrNilStatusCoreComponents signals that a nil status core components has been provided
var ErrNilStatusCoreComponents = errors.New("nil status core components")

// ErrNilCryptoComponents signals that a nil crypto components instance has been provided
var ErrNilCryptoComponents = errors.New("nil crypto components")

// ErrNilDataComponents signals that a nil data components instance has been provided
var ErrNilDataComponents = errors.New("nil data components")

// ErrTransactionValueLengthTooBig signals that a too big value has been given to a transaction
var ErrTransactionValueLengthTooBig = errors.New("value length is too big")

// ErrNilNetworkComponents signals that a nil network components instance has been provided
var ErrNilNetworkComponents = errors.New("nil network components")

// ErrNilProcessComponents signals that a nil process components instance has been provided
var ErrNilProcessComponents = errors.New("nil process components")

// ErrNilStateComponents signals that a nil state components instance has been provided
var ErrNilStateComponents = errors.New("nil state components")

// ErrNilStatusComponents signals that a nil status components instance has been provided
var ErrNilStatusComponents = errors.New("nil status components")

// ErrNodeCloseFailed signals that the close function of the node failed
var ErrNodeCloseFailed = errors.New("node closing failed ")

// ErrDifferentSenderShardId signals that a different shard ID was detected between the sender shard ID and the current node shard ID
var ErrDifferentSenderShardId = errors.New("different shard ID between the transaction sender shard ID and current node shard ID")

// ErrInvalidESDTRole signals that an invalid ESDT role has been provided
var ErrInvalidESDTRole = errors.New("invalid ESDT role")

// ErrMetachainOnlyEndpoint signals that an endpoint was called, but it is only available for metachain nodes
var ErrMetachainOnlyEndpoint = errors.New("the endpoint is only available on metachain nodes")

// ErrCannotCastAccountHandlerToUserAccountHandler signals that an account handler cannot be cast to user account handler
var ErrCannotCastAccountHandlerToUserAccountHandler = errors.New("cannot cast account handler to user account handler")

// ErrCannotCastUserAccountHandlerToVmCommonUserAccountHandler signals that an user account handler cannot be cast to vm common user account handler
var ErrCannotCastUserAccountHandlerToVmCommonUserAccountHandler = errors.New("cannot cast user account handler to vm common user account handler")

// ErrTrieOperationsTimeout signals that a trie operation took too long
var ErrTrieOperationsTimeout = errors.New("trie operations timeout")

// ErrNilStatusHandler signals that a nil status handler was provided
var ErrNilStatusHandler = errors.New("nil status handler")

// ErrNilCreateTransactionArgs signals that create transaction args is nil
var ErrNilCreateTransactionArgs = errors.New("nil args for create transaction")

// ErrNilComponentHandler signals that a nil component handler was provided
var ErrNilComponentHandler = errors.New("nil component handler")

// ErrDuplicatedComponentHandler signals that a duplicated component handler was provided
var ErrDuplicatedComponentHandler = errors.New("duplicated component handler")

// ErrInvalidComponentHandler signals that an invalid component handler was provided
var ErrInvalidComponentHandler = errors.New("invalid component handler")
