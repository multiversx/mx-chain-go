package shardchain

import (
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/storage"
	"sync"
)

// ArgsNewShardEndOfEpochTrigger defines the arguments needed for new end of epoch trigger
type ArgsNewShardEndOfEpochTrigger struct {
}

type trigger struct {
	epoch             uint32
	currentRoundIndex int64
	epochStartRound   int64
	isEndOfEpoch      bool
	finality          uint64

	newEpochHdrReceived bool

	mutReceived    sync.Mutex
	mapHashHdr     map[string]*block.MetaBlock
	mapNonceHashes map[uint64][]string

	metaHdrPool    storage.Cacher
	metaHdrStorage storage.Storer

	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher
}

// NewEndOfEpochTrigger creates a trigger to signal end of epoch
func NewEndOfEpochTrigger(args *ArgsNewShardEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewShardEndOfEpochTrigger
	}

	return &trigger{}, nil
}

// IsEndOfEpoch returns true if conditions are fullfilled for end of epoch
func (t *trigger) IsEndOfEpoch() bool {
	return t.isEndOfEpoch
}

// Epoch returns the current epoch number
func (t *trigger) Epoch() uint32 {
	return t.epoch
}

// ForceEndOfEpoch sets the conditions for end of epoch to true in case of edge cases
func (t *trigger) ForceEndOfEpoch(round int64) error {
	return nil
}

// ReceivedHeader saves the header into pool to verify if end-of-epoch conditions are fullfilled
func (t *trigger) ReceivedHeader(header data.HeaderHandler) {
	metaHdr, ok := header.(*block.MetaBlock)
	if !ok {
		return
	}

	t.mutReceived.Lock()
	defer t.mutReceived.Unlock()

	hdrHash, err := core.CalculateHash(t.marshalizer, t.hasher, metaHdr)
	if err != nil {
		return
	}

	if _, ok := t.mapHashHdr[string(hdrHash)]; ok {
		return
	}

	if len(metaHdr.EndOfEpoch.LastFinalizedHeaders) > 0 {

	}
}

// Update updates the end-of-epoch trigger
func (t *trigger) Update(round int64) {
}

// Processed signals that end-of-epoch has been processed
func (t *trigger) Processed() {
	t.isEndOfEpoch = false
	t.newEpochHdrReceived = false

	//TODO: do a cleanup
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
