package transaction

import (
    "bytes"
    "math/big"

    "github.com/ElrondNetwork/elrond-go/crypto"
    "github.com/ElrondNetwork/elrond-go/data/state"
    "github.com/ElrondNetwork/elrond-go/data/transaction"
    "github.com/ElrondNetwork/elrond-go/hashing"
    "github.com/ElrondNetwork/elrond-go/marshal"
    "github.com/ElrondNetwork/elrond-go/process"
    "github.com/ElrondNetwork/elrond-go/sharding"
)

// InterceptedTransaction holds and manages a transaction based struct with extended functionality
type InterceptedTransaction struct {
    tx                       *transaction.Transaction
    marshalizer              marshal.Marshalizer
    hasher                   hashing.Hasher
    keyGen                   crypto.KeyGenerator
    singleSigner             crypto.SingleSigner
    addrConv                 state.AddressConverter
    coordinator              sharding.Coordinator
    hash                     []byte
    rcvShard                 uint32
    sndShard                 uint32
    isAddressedToOtherShards bool
}

// NewInterceptedTransaction returns a new instance of InterceptedTransaction
func NewInterceptedTransaction(
    txBuff []byte,
    marshalizer marshal.Marshalizer,
    hasher hashing.Hasher,
    keyGen crypto.KeyGenerator,
    signer crypto.SingleSigner,
    addrConv state.AddressConverter,
    coordinator sharding.Coordinator,
) (*InterceptedTransaction, error) {

    if txBuff == nil {
        return nil, process.ErrNilBuffer
    }
    if marshalizer == nil || marshalizer.IsInterfaceNil() {
        return nil, process.ErrNilMarshalizer
    }
    if hasher == nil || hasher.IsInterfaceNil() {
        return nil, process.ErrNilHasher
    }
    if keyGen == nil || keyGen.IsInterfaceNil() {
        return nil, process.ErrNilKeyGen
    }
    if signer == nil || signer.IsInterfaceNil() {
        return nil, process.ErrNilSingleSigner
    }
    if addrConv == nil || addrConv.IsInterfaceNil() {
        return nil, process.ErrNilAddressConverter
    }
    if coordinator == nil || coordinator.IsInterfaceNil() {
        return nil, process.ErrNilShardCoordinator
    }

    tx := &transaction.Transaction{}
    err := marshalizer.Unmarshal(tx, txBuff)
    if err != nil {
        return nil, err
    }

    inTx := &InterceptedTransaction{
        tx:           tx,
        marshalizer:  marshalizer,
        hasher:       hasher,
        singleSigner: signer,
        addrConv:     addrConv,
        keyGen:       keyGen,
        coordinator:  coordinator,
    }

    txBuffWithoutSig, err := inTx.processFields(txBuff)
    if err != nil {
        return nil, err
    }

    err = inTx.integrity()
    if err != nil {
        return nil, err
    }

    err = inTx.verifySig(txBuffWithoutSig)
    if err != nil {
        return nil, err
    }

    return inTx, nil
}

func (inTx *InterceptedTransaction) processFields(txBuffWithSig []byte) ([]byte, error) {
    copiedTx := *inTx.Transaction()
    copiedTx.Signature = nil
    buffCopiedTx, err := inTx.marshalizer.Marshal(&copiedTx)
    if err != nil {
        return nil, err
    }
    inTx.hash = inTx.hasher.Compute(string(txBuffWithSig))

    sndAddr, err := inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.tx.SndAddr)
    if err != nil {
        return nil, process.ErrInvalidSndAddr
    }

    rcvAddr, err := inTx.addrConv.CreateAddressFromPublicKeyBytes(inTx.tx.RcvAddr)
    if err != nil {
        return nil, process.ErrInvalidRcvAddr
    }

    inTx.sndShard = inTx.coordinator.ComputeId(sndAddr)
    emptyAddr := make([]byte, len(rcvAddr.Bytes()))
    inTx.rcvShard = inTx.coordinator.ComputeId(rcvAddr)
    if bytes.Equal(rcvAddr.Bytes(), emptyAddr) {
        inTx.rcvShard = inTx.sndShard
    }

    inTx.isAddressedToOtherShards = inTx.rcvShard != inTx.coordinator.SelfId() &&
        inTx.sndShard != inTx.coordinator.SelfId()

    return buffCopiedTx, nil
}

// integrity checks for not nil fields and negative value
func (inTx *InterceptedTransaction) integrity() error {
    if inTx.tx.Signature == nil {
        return process.ErrNilSignature
    }

    if inTx.tx.RcvAddr == nil {
        return process.ErrNilRcvAddr
    }

    if inTx.tx.SndAddr == nil {
        return process.ErrNilSndAddr
    }

    if inTx.tx.Value == nil {
        return process.ErrNilValue
    }

    if inTx.tx.Value.Cmp(big.NewInt(0)) < 0 {
        return process.ErrNegativeValue
    }

    return nil
}

// verifySig checks if the tx is correctly signed
func (inTx *InterceptedTransaction) verifySig(txBuffWithoutSig []byte) error {
    senderPubKey, err := inTx.keyGen.PublicKeyFromByteArray(inTx.tx.SndAddr)
    if err != nil {
        return err
    }

    err = inTx.singleSigner.Verify(senderPubKey, txBuffWithoutSig, inTx.tx.Signature)
    if err != nil {
        return err
    }

    return nil
}

// RcvShard returns the receiver shard
func (inTx *InterceptedTransaction) RcvShard() uint32 {
    return inTx.rcvShard
}

// SndShard returns the sender shard
func (inTx *InterceptedTransaction) SndShard() uint32 {
    return inTx.sndShard
}

// IsAddressedToOtherShards returns true if this transaction is not meant to be processed by the node from this shard
func (inTx *InterceptedTransaction) IsAddressedToOtherShards() bool {
    return inTx.isAddressedToOtherShards
}

// Transaction returns the transaction pointer that actually holds the data
func (inTx *InterceptedTransaction) Transaction() *transaction.Transaction {
    return inTx.tx
}

// Hash gets the hash of this transaction
func (inTx *InterceptedTransaction) Hash() []byte {
    return inTx.hash
}
