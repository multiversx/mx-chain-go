package interceptors

import (
	"errors"
	"strings"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/core/sync"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/storage"
)

type interceptedDataStatus int8

const (
	validInterceptedData           interceptedDataStatus = iota
	interceptedDataStatusBytesSize                       = 8
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
func (idv *interceptedDataVerifier) Verify(interceptedData process.InterceptedData, topic string) error {
	if len(interceptedData.Hash()) == 0 {
		return interceptedData.CheckValidity()
	}

	exists, err := idv.checkCachedData(interceptedData, topic)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	err = interceptedData.CheckValidity()
	if err != nil {
		logInterceptedDataCheckValidityErr(interceptedData, err)
		// TODO: investigate to selectively add as invalid intercepted data only when data is indeed invalid instead of missing
		// idv.cache.Put(interceptedData.Hash(), invalidInterceptedData, interceptedDataStatusBytesSize)
		return process.ErrInvalidInterceptedData
	}

	return nil
}

// MarkVerified marks the intercepted data as verified
func (idv *interceptedDataVerifier) MarkVerified(interceptedData process.InterceptedData, topic string) {
	if !isCrossShardTopic(topic) {
		return
	}

	if len(interceptedData.Hash()) == 0 {
		return
	}

	idv.km.Lock(string(interceptedData.Hash()))
	defer idv.km.Unlock(string(interceptedData.Hash()))

	idv.cache.Put(interceptedData.Hash(), validInterceptedData, interceptedDataStatusBytesSize)
}

func (idv *interceptedDataVerifier) checkCachedData(interceptedData process.InterceptedData, topic string) (bool, error) {
	hash := string(interceptedData.Hash())
	idv.km.RLock(hash)
	defer idv.km.RUnlock(hash)

	val, ok := idv.cache.Get(interceptedData.Hash())
	if !ok {
		return ok, nil
	}

	if val != validInterceptedData {
		return ok, process.ErrInvalidInterceptedData
	}

	if !isCrossShardTopic(topic) {
		return ok, nil
	}

	if !interceptedData.ShouldAllowDuplicates() {
		return ok, process.ErrDuplicatedInterceptedDataNotAllowed
	}

	return ok, nil
}

func isCrossShardTopic(topic string) bool {
	baseTopic := strings.Trim(topic, core.TopicRequestSuffix)
	topicSplit := strings.Split(baseTopic, "_")
	return len(topicSplit) == 3 || len(topicSplit) == 1 // cross _0_1 or global topic
}

func logInterceptedDataCheckValidityErr(interceptedData process.InterceptedData, err error) {
	if errors.Is(err, common.ErrAlreadyExistingEquivalentProof) {
		log.Trace("Intercepted data is invalid", "hash", interceptedData.Hash(), "err", err)
		return
	}

	log.Debug("Intercepted data is invalid", "hash", interceptedData.Hash(), "err", err)
}

// IsInterfaceNil returns true if there is no value under the interface
func (idv *interceptedDataVerifier) IsInterfaceNil() bool {
	return idv == nil
}
