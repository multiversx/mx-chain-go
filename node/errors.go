package node

import (
	"errors"
)

// ErrNilMarshalizer signals that a nil marshalizer has been provided
var ErrNilMarshalizer = errors.New("trying to set nil marshalizer")

// ErrNilMessenger signals that a nil messenger has been provided
var ErrNilMessenger = errors.New("nil messenger")

// ErrNilContext signals that a nil context has been provided
var ErrNilContext = errors.New("trying to set nil context")

// ErrNilHasher signals that a nil hasher has been provided
var ErrNilHasher = errors.New("trying to set nil hasher")

// ErrNilAccountsAdapter signals that a nil accounts adapter has been provided
var ErrNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

// ErrNilAddressConverter signals that a nil address converter has been provided
var ErrNilAddressConverter = errors.New("trying to set nil address converter")

// ErrNilBlockchain signals that a nil blockchain structure has been provided
var ErrNilBlockchain = errors.New("nil blockchain")

// ErrNilPrivateKey signals that a nil private key has been provided
var ErrNilPrivateKey = errors.New("trying to set nil private key")

// ErrNilSingleSignKeyGen signals that a nil single key generator has been provided
var ErrNilSingleSignKeyGen = errors.New("trying to set nil single sign key generator")

// ErrNilPublicKey signals that a nil public key has been provided
var ErrNilPublicKey = errors.New("trying to set nil public key")

// ErrZeroRoundDurationNotSupported signals that 0 seconds round duration is not supported
var ErrZeroRoundDurationNotSupported = errors.New("0 round duration time is not supported")

// ErrNegativeOrZeroConsensusGroupSize signals that 0 elements consensus group is not supported
var ErrNegativeOrZeroConsensusGroupSize = errors.New("group size should be a strict positive number")

// ErrNilSyncTimer signals that a nil sync timer has been provided
var ErrNilSyncTimer = errors.New("trying to set nil sync timer")

// ErrNilBlockProcessor signals that a nil block processor has been provided
var ErrNilBlockProcessor = errors.New("trying to set nil block processor")

// ErrNilDataPool signals that a nil data pool has been provided
var ErrNilDataPool = errors.New("trying to set nil data pool")

// ErrNilShardCoordinator signals that a nil shard coordinator has been provided
var ErrNilShardCoordinator = errors.New("trying to set nil shard coordinator")

// ErrNilUint64ByteSliceConverter signals that a nil uint64 <-> byte slice converter has been provided
var ErrNilUint64ByteSliceConverter = errors.New("trying to set nil uint64 - byte slice converter")

// ErrNilBalances signals that a nil list of initial balances has been provided
var ErrNilBalances = errors.New("trying to set nil balances")

// ErrNilMultiSig signals that a nil multisig object has been provided
var ErrNilMultiSig = errors.New("trying to set nil multisig")

// ErrNilForkDetector signals that a nil forkdetector object has been provided
var ErrNilForkDetector = errors.New("nil fork detector")
