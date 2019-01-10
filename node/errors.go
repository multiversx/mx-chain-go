package node

import (
	"errors"
)

var errNodeNotStarted = errors.New("node is not started yet")

var errNilMarshalizer = errors.New("trying to set nil marshalizer")

var errNilContext = errors.New("trying to set nil context")

var errNilHasher = errors.New("trying to set nil hasher")

var errNilAccountsAdapter = errors.New("trying to set nil accounts adapter")

var errNilAddressConverter = errors.New("trying to set nil address converter")

var errNilBlockchain = errors.New("trying to set nil blockchain")

var errNilPrivateKey = errors.New("trying to set nil private key")

var errNilSingleSignKeyGen = errors.New("trying to set nil single sign key generator")

var errNilPublicKey = errors.New("trying to set nil public key")

var errZeroRoundDurationNotSupported = errors.New("0 round duration time is not supported")

var errNegativeOrZeroConsensusGroupSize = errors.New("group size should be a strict positive number")

var errNilSyncTimer = errors.New("trying to set nil sync timer")

var errNilBlockProcessor = errors.New("trying to set nil block processor")

var errNilDataPool = errors.New("trying to set nil data pool")

var errNilShardCoordinator = errors.New("trying to set nil shard coordinator")

var errNilUint64ByteSliceConverter = errors.New("trying to set nil uint64 - byte slice converter")

var errNilBalances = errors.New("trying to set nil balances")

var errNilMultiSig = errors.New("trying to set nil multisig")
