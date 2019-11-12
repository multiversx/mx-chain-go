package shardchain

import (
	"bytes"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/data/typeConverters"
	"github.com/ElrondNetwork/elrond-go/dataRetriever"
	"github.com/ElrondNetwork/elrond-go/endOfEpoch"
	"github.com/ElrondNetwork/elrond-go/hashing"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/ElrondNetwork/elrond-go/storage"
	"sync"
)

// ArgsShardEndOfEpochTrigger struct { defines the arguments needed for new end of epoch trigger
type ArgsShardEndOfEpochTrigger struct {
	Marshalizer marshal.Marshalizer
	Hasher      hashing.Hasher

	HeaderValidator endOfEpoch.HeaderValidator
	Uint64Converter typeConverters.Uint64ByteSliceConverter

	DataPool       dataRetriever.PoolsHolder
	Storage        dataRetriever.StorageService
	RequestHandler endOfEpoch.RequestHandler

	Epoch    uint32
	Validity uint64
	Finality uint64
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

	metaHdrPool         storage.Cacher
	metaHdrNonces       dataRetriever.Uint64SyncMapCacher
	metaHdrStorage      storage.Storer
	metaNonceHdrStorage storage.Storer
	uint64Converter     typeConverters.Uint64ByteSliceConverter

	marshalizer     marshal.Marshalizer
	hasher          hashing.Hasher
	headerValidator endOfEpoch.HeaderValidator

	requestHandler endOfEpoch.RequestHandler
}

// NewEndOfEpochTrigger creates a trigger to signal end of epoch
func NewEndOfEpochTrigger(args *ArgsShardEndOfEpochTrigger) (*trigger, error) {
	if args == nil {
		return nil, endOfEpoch.ErrNilArgsNewShardEndOfEpochTrigger
	}
	if check.IfNil(args.Hasher) {
		return nil, endOfEpoch.ErrNilHasher
	}
	if check.IfNil(args.Marshalizer) {
		return nil, endOfEpoch.ErrNilMarshalizer
	}
	if check.IfNil(args.HeaderValidator) {
		return nil, endOfEpoch.ErrNilHeaderValidator
	}
	if check.IfNil(args.DataPool) {
		return nil, endOfEpoch.ErrNilDataPoolsHolder
	}
	if check.IfNil(args.Storage) {
		return nil, endOfEpoch.ErrNilStorageService
	}
	if check.IfNil(args.RequestHandler) {
		return nil, endOfEpoch.ErrNilRequestHandler
	}
	if check.IfNil(args.DataPool.MetaBlocks()) {
		return nil, endOfEpoch.ErrNilMetaBlocksPool
	}
	if check.IfNil(args.DataPool.HeadersNonces()) {
		return nil, endOfEpoch.ErrNilHeaderNoncesPool
	}
	if check.IfNil(args.Uint64Converter) {
		return nil, endOfEpoch.ErrNilUint64Converter
	}

	metaHdrStorage := args.Storage.GetStorer(dataRetriever.MetaBlockUnit)
	if check.IfNil(metaHdrStorage) {
		return nil, endOfEpoch.ErrNilMetaHdrStorage
	}

	metaHdrNoncesStorage := args.Storage.GetStorer(dataRetriever.MetaHdrNonceHashDataUnit)
	if check.IfNil(metaHdrNoncesStorage) {
		return nil, endOfEpoch.ErrNilMetaNonceHashStorage
	}

	newTrigger := &trigger{
		epoch:               args.Epoch,
		currentRoundIndex:   0,
		epochStartRound:     0,
		isEndOfEpoch:        false,
		validity:            args.Validity,
		finality:            args.Finality,
		newEpochHdrReceived: false,
		mutReceived:         sync.Mutex{},
		mapHashHdr:          make(map[string]*block.MetaBlock),
		mapNonceHashes:      make(map[uint64][]string),
		mapEndOfEpochHdrs:   make(map[string]*block.MetaBlock),
		metaHdrPool:         args.DataPool.MetaBlocks(),
		metaHdrNonces:       args.DataPool.HeadersNonces(),
		metaHdrStorage:      metaHdrStorage,
		metaNonceHdrStorage: metaHdrNoncesStorage,
		uint64Converter:     args.Uint64Converter,
		marshalizer:         args.Marshalizer,
		hasher:              args.Hasher,
		headerValidator:     args.HeaderValidator,
		requestHandler:      args.RequestHandler,
	}
	return newTrigger, nil
}

// IsEndOfEpoch returns true if conditions are fulfilled for end of epoch
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

// ReceivedHeader saves the header into pool to verify if end-of-epoch conditions are fulfilled
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

func (t *trigger) checkIfTriggerCanBeActivated(hash string, metaHdr *block.MetaBlock) bool {
	isMetaHdrValid := true
	currHdr := metaHdr
	for i := metaHdr.Nonce - 1; i >= metaHdr.Nonce-t.validity; i-- {
		neededHdr, err := t.getHeaderWithNonceAndHash(i, currHdr.PrevHash)
		if err != nil {
			isMetaHdrValid = false
		}

		err = t.headerValidator.IsHeaderConstructionValid(currHdr, neededHdr)
		if err != nil {
			isMetaHdrValid = false
		}
	}

	isMetaHdrFinal := false
	nextBlocksVerified := uint64(0)
	currHdr = metaHdr
	for i := metaHdr.Nonce + 1; i <= metaHdr.Nonce+t.finality; i++ {
		currHash, err := core.CalculateHash(t.marshalizer, t.hasher, currHdr)
		if err != nil {
			continue
		}

		neededHdr, err := t.getHeaderWithNonceAndPrevHash(i, currHash)

		err = t.headerValidator.IsHeaderConstructionValid(neededHdr, currHdr)
		if err != nil {
			continue
		}

		currHdr = neededHdr
		nextBlocksVerified += 1
	}

	if nextBlocksVerified < t.finality {
		for i := currHdr.Nonce + 1; i <= currHdr.Nonce+t.finality; i++ {
			go t.requestHandler.RequestHeaderByNonce(sharding.MetachainShardId, i)
		}
	} else {
		isMetaHdrFinal = true
	}

	return isMetaHdrFinal && isMetaHdrValid
}

func (t *trigger) getHeaderWithNonceAndHash(nonce uint64, neededHash []byte) (*block.MetaBlock, error) {
	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		if !bytes.Equal(neededHash, []byte(hash)) {
			continue
		}

		neededHdr := t.mapHashHdr[hash]
		if neededHdr != nil {
			return neededHdr, nil
		}
	}

	peekedData, _ := t.metaHdrPool.Peek(neededHash)
	neededHdr, ok := peekedData.(*block.MetaBlock)
	if ok {
		t.mapHashHdr[string(neededHash)] = neededHdr
		t.mapNonceHashes[nonce] = append(t.mapNonceHashes[nonce], string(neededHash))
		return neededHdr, nil
	}

	storageData, err := t.metaHdrStorage.Get(neededHash)
	if err == nil {
		var neededHdr block.MetaBlock
		err = t.marshalizer.Unmarshal(&neededHdr, storageData)
		if err == nil {
			t.mapHashHdr[string(neededHash)] = &neededHdr
			t.mapNonceHashes[nonce] = append(t.mapNonceHashes[nonce], string(neededHash))
			return &neededHdr, nil
		}
	}

	go t.requestHandler.RequestHeader(sharding.MetachainShardId, neededHash)

	return nil, endOfEpoch.ErrMetaHdrNotFound
}

func (t *trigger) getHeaderWithNonceAndPrevHash(nonce uint64, prevHash []byte) (*block.MetaBlock, error) {
	metaHdrHashesWithNonce := t.mapNonceHashes[nonce]
	for _, hash := range metaHdrHashesWithNonce {
		hdrWithNonce := t.mapHashHdr[hash]
		if hdrWithNonce != nil && bytes.Equal(hdrWithNonce.PrevHash, prevHash) {
			return hdrWithNonce, nil
		}
	}

	shIdMap, ok := t.metaHdrNonces.Get(nonce)
	if ok {
		hdrHash, ok := shIdMap.Load(sharding.MetachainShardId)
		if ok {
			dataHdr, _ := t.metaHdrPool.Peek(hdrHash)
			hdrWithNonce, ok := dataHdr.(*block.MetaBlock)
			if ok && bytes.Equal(hdrWithNonce.PrevHash, prevHash) {
				return hdrWithNonce, nil
			}
		}
	}

	nonceToByteSlice := t.uint64Converter.ToByteSlice(nonce)
	dataHdr, err := t.metaNonceHdrStorage.Get(nonceToByteSlice)
	if err != nil || len(dataHdr) == 0 {
		go t.requestHandler.RequestHeaderByNonce(sharding.MetachainShardId, nonce)
		return nil, err
	}

	var neededHash []byte
	err = t.marshalizer.Unmarshal(neededHash, dataHdr)
	if err == nil {
		return nil, err
	}

	return t.getHeaderWithNonceAndHash(nonce, neededHash)
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
