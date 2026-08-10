package interceptors

import (
	"errors"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/p2p"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type interceptedDataStatus int8

const (
	validInterceptedData interceptedDataStatus = iota
	pendingValidInterceptedData
	interceptedDataStatusBytesSize = 8
)

type interceptedDataVerifier struct {
	km    sync.KeyRWMutexHandler
	cache storage.Cacher
}

// NewInterceptedDataVerifier creates a new instance of intercepted data verifier
func NewInterceptedDataVerifier(cache storage.Cacher) (*interceptedDataVerifier, error) {
	if check.IfNil(cache) {
		return nil, process.ErrNilInterceptedDataCache
	}

	return &interceptedDataVerifier{
		km:    sync.NewKeyRWMutex(),
		cache: cache,
	}, nil
}

// Verify will check if the intercepted data has been validated before and put in the time cache.
// It will retrieve the status in the cache if it exists, otherwise it will validate it and store the status of the
// validation in the cache. Note that the entries are stored for a set period of time
func (idv *interceptedDataVerifier) Verify(interceptedData process.InterceptedData, topic string, broadcastMethod p2p.BroadcastMethod) error {
	if len(interceptedData.Hash()) == 0 {
		return interceptedData.CheckValidity()
	}

	hash := string(interceptedData.Hash())
	idv.km.RLock(hash)
	val, ok := idv.cache.Get(interceptedData.Hash())
	idv.km.RUnlock(hash)

	if ok {
		err := checkCachedData(val, topic, broadcastMethod, interceptedData)
		if err != nil {
			return err
		}

		return nil
	}

	// Validate and mark with write lock to prevent race condition
	idv.km.Lock(hash)
	defer idv.km.Unlock(hash)

	// Double-check cache after acquiring write lock
	val, ok = idv.cache.Get(interceptedData.Hash())
	if ok {
		err := checkCachedData(val, topic, broadcastMethod, interceptedData)
		if err != nil {
			return err
		}

		return nil
	}

	// Validate the data
	err := interceptedData.CheckValidity()
	if err != nil {
		logInterceptedDataCheckValidityErr(interceptedData, err)
		// TODO: investigate to selectively add as invalid intercepted data only when data is indeed invalid instead of missing
		// idv.cache.Put(interceptedData.Hash(), invalidInterceptedData, interceptedDataStatusBytesSize)
		return process.ErrInvalidInterceptedData
	}

	// Mark as verified pending immediately after successful validation
	if !isDirectSend(broadcastMethod) {
		idv.cache.Put(interceptedData.Hash(), pendingValidInterceptedData, interceptedDataStatusBytesSize)
	}

	return nil
}

// MarkVerified marks the intercepted data as verified
func (idv *interceptedDataVerifier) MarkVerified(interceptedData process.InterceptedData, broadcastMethod p2p.BroadcastMethod) {
	if len(interceptedData.Hash()) == 0 {
		return
	}
	if isDirectSend(broadcastMethod) {
		return
	}

	idv.km.Lock(string(interceptedData.Hash()))
	defer idv.km.Unlock(string(interceptedData.Hash()))

	idv.cache.Put(interceptedData.Hash(), validInterceptedData, interceptedDataStatusBytesSize)
}

func isDirectSend(broadcastMethod p2p.BroadcastMethod) bool {
	return broadcastMethod == p2p.Direct
}

func shouldCheckForDuplicates(
	topic string,
	broadcastMethod p2p.BroadcastMethod,
	cachedValue interface{},
) bool {
	if isDirectSend(broadcastMethod) {
		return false // skip deduplication on direct messages
	}

	// if the value was not yet marked as verified, do not check for duplicates yet
	isPending := cachedValue == pendingValidInterceptedData

	topicSplit := strings.Split(topic, "_")
	isCrossShardTopic := len(topicSplit) == 3 || len(topicSplit) == 1 // cross _0_1 or global topic
	return isCrossShardTopic && !isPending
}

func logInterceptedDataCheckValidityErr(interceptedData process.InterceptedData, err error) {
	if errors.Is(err, common.ErrAlreadyExistingEquivalentProof) {
		log.Trace("Intercepted data is invalid", "hash", interceptedData.Hash(), "err", err)
		return
	}

	log.Debug("Intercepted data is invalid", "hash", interceptedData.Hash(), "err", err)
}

func checkCachedData(
	cachedValue interface{},
	topic string,
	broadcastMethod p2p.BroadcastMethod,
	interceptedData process.InterceptedData,
) error {
	if !isValidInterceptedData(cachedValue) {
		return process.ErrInvalidInterceptedData
	}

	if !shouldCheckForDuplicates(topic, broadcastMethod, cachedValue) {
		return nil
	}

	if !interceptedData.ShouldAllowDuplicates() {
		return process.ErrDuplicatedInterceptedDataNotAllowed
	}

	return nil
}

func isValidInterceptedData(val interface{}) bool {
	return val == validInterceptedData || val == pendingValidInterceptedData
}

// IsInterfaceNil returns true if there is no value under the interface
func (idv *interceptedDataVerifier) IsInterfaceNil() bool {
	return idv == nil
}
