package shardchain

import (
	"bytes"
	"fmt"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"github.com/prometheus/common/log"
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
	validity          uint64

	newEpochHdrReceived bool

	mutReceived       sync.Mutex
	mapHashHdr        map[string]*block.MetaBlock
	mapNonceHashes    map[uint64][]string
	mapEndOfEpochHdrs map[string]*block.MetaBlock

	metaHdrPool    storage.Cacher
	metaHdrNonces  dataRetriever.Uint64SyncMapCacher
	metaHdrStorage storage.Storer

	marshalizer marshal.Marshalizer
	hasher      hashing.Hasher

	onRequestHeaderHandlerByNonce func(shardId uint32, nonce uint64)
	onRequestHeaderHandler        func(shardId uint32, hash []byte)
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
	if t.isEndOfEpoch == true {
		return
	}

	metaHdr, ok := header.(*block.MetaBlock)
	if !ok {
		return
	}

	if t.newEpochHdrReceived == false && len(metaHdr.EndOfEpoch.LastFinalizedHeaders) == 0 {
		return
	}

	hdrHash, err := core.CalculateHash(t.marshalizer, t.hasher, metaHdr)
	if err != nil {
		return
	}

	t.mutReceived.Lock()

	if len(metaHdr.EndOfEpoch.LastFinalizedHeaders) > 0 {
		t.newEpochHdrReceived = true
		t.mapEndOfEpochHdrs[string(hdrHash)] = metaHdr
	} else {
		t.mapHashHdr[string(hdrHash)] = metaHdr
		t.mapNonceHashes[metaHdr.Nonce] = append(t.mapNonceHashes[metaHdr.Nonce], string(hdrHash))
	}

	for hash, meta := range t.mapEndOfEpochHdrs {
		t.isEndOfEpoch = t.checkIfTriggerCanBeActivated(hash, meta)
	}

	t.mutReceived.Unlock()
}

func (t *trigger) getHeader(nonce uint64, neededHash []byte) (*block.MetaBlock, error) {
	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		if bytes.Equal(neededHash, []byte(hash)) {
			neededHdr := t.mapHashHdr[hash]
			if neededHdr != nil {
				return neededHdr, nil
			}
		}
	}

	peekedData, ok := t.metaHdrPool.Peek(neededHash)
	if ok {
		neededHdr, ok := peekedData.(*block.MetaBlock)
		if ok {
			t.mapHashHdr[string(neededHash)] = neededHdr
			t.mapNonceHashes[nonce] = append(t.mapNonceHashes[nonce], string(neededHash))
			return neededHdr, nil
		}
	}

	storageData, err := t.metaHdrStorage.Get(neededHash)
	if err != nil {
		var neededHdr block.MetaBlock
		err = t.marshalizer.Unmarshal(&neededHdr, storageData)
		if err != nil {
			t.mapHashHdr[string(neededHash)] = &neededHdr
			t.mapNonceHashes[nonce] = append(t.mapNonceHashes[nonce], string(neededHash))
			return &neededHdr, nil
		}
	}

	go t.onRequestHeaderHandler(sharding.MetachainShardId, neededHash)

	return nil, endOfEpoch.ErrMetaHdrNotFound
}

func (t *trigger) checkIfTriggerCanBeActivated(hash string, metaHdr *block.MetaBlock) bool {
	isMetaHdrValid := true
	currHdr := metaHdr
	for i := metaHdr.Nonce - 1; i >= metaHdr.Nonce-t.validity; i-- {
		neededHdr, err := t.getHeader(i, currHdr.PrevHash)
		if err != nil {
			isMetaHdrValid = false
		}

		err = t.isHdrConstructionValid(currHdr, neededHdr)
		if err != nil {
			isMetaHdrValid = false
		}
	}

	for i := metaHdr.Nonce + 1; i <= metaHdr.Nonce+t.finality; i++ {

	}

	isMetaHdrFinal := true
	// verify if there are "K" block after current to make this one final
	nextBlocksVerified := uint32(0)
	for _, shardHdr := range finalityAttestingShardHdrs[shardId] {
		if nextBlocksVerified >= mp.shardBlockFinality {
			break
		}

		// found a header with the next nonce
		if shardHdr.GetNonce() == lastVerifiedHdr.GetNonce()+1 {
			err := mp.isHdrConstructionValid(shardHdr, lastVerifiedHdr)
			if err != nil {
				log.Debug(err.Error())
				continue
			}

			lastVerifiedHdr = shardHdr
			nextBlocksVerified += 1
		}
	}

	if nextBlocksVerified < mp.shardBlockFinality {
		go mp.onRequestHeaderHandlerByNonce(lastVerifiedHdr.GetShardID(), lastVerifiedHdr.GetNonce()+1)
		return process.ErrHeaderNotFinal
	}

	return isMetaHdrFinal && isMetaHdrValid
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

func (t *trigger) isHdrConstructionValid(currHdr, prevHdr data.HeaderHandler) error {
	if prevHdr == nil || prevHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}
	if currHdr == nil || currHdr.IsInterfaceNil() {
		return process.ErrNilBlockHeader
	}

	// special case with genesis nonce - 0
	if currHdr.GetNonce() == 0 {
		if prevHdr.GetNonce() != 0 {
			return process.ErrWrongNonceInBlock
		}
		// block with nonce 0 was already saved
		if prevHdr.GetRootHash() != nil {
			return process.ErrRootStateDoesNotMatch
		}
		return nil
	}

	//TODO: add verification if rand seed was correctly computed add other verification
	//TODO: check here if the 2 header blocks were correctly signed and the consensus group was correctly elected
	if prevHdr.GetRound() >= currHdr.GetRound() {
		log.Debug(fmt.Sprintf("round does not match in shard %d: local block round is %d and node received block with round %d\n",
			currHdr.GetShardID(), prevHdr.GetRound(), currHdr.GetRound()))
		return process.ErrLowerRoundInBlock
	}

	if currHdr.GetNonce() != prevHdr.GetNonce()+1 {
		log.Debug(fmt.Sprintf("nonce does not match in shard %d: local block nonce is %d and node received block with nonce %d\n",
			currHdr.GetShardID(), prevHdr.GetNonce(), currHdr.GetNonce()))
		return process.ErrWrongNonceInBlock
	}

	prevHeaderHash, err := core.CalculateHash(t.marshalizer, t.hasher, prevHdr)
	if err != nil {
		return err
	}

	if !bytes.Equal(currHdr.GetPrevHash(), prevHeaderHash) {
		log.Debug(fmt.Sprintf("block hash does not match in shard %d: local block hash is %s and node received block with previous hash %s\n",
			currHdr.GetShardID(), core.ToB64(prevHeaderHash), core.ToB64(currHdr.GetPrevHash())))
		return process.ErrBlockHashDoesNotMatch
	}

	if !bytes.Equal(currHdr.GetPrevRandSeed(), prevHdr.GetRandSeed()) {
		log.Debug(fmt.Sprintf("random seed does not match in shard %d: local block random seed is %s and node received block with previous random seed %s\n",
			currHdr.GetShardID(), core.ToB64(prevHdr.GetRandSeed()), core.ToB64(currHdr.GetPrevRandSeed())))
		return process.ErrRandSeedDoesNotMatch
	}

	return nil
}

// IsInterfaceNil returns true if underlying object is nil
func (t *trigger) IsInterfaceNil() bool {
	return t == nil
}
