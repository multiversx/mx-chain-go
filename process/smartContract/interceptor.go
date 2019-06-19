package smartContract

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/data/state"
	"github.com/ElrondNetwork/elrond-go-sandbox/dataRetriever"
	"github.com/ElrondNetwork/elrond-go-sandbox/hashing"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/p2p"
	"github.com/ElrondNetwork/elrond-go-sandbox/process"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
	"github.com/ElrondNetwork/elrond-go-sandbox/storage"
)

// ScrInterceptor is used for intercepting transaction and storing them into a datapool
type ScrInterceptor struct {
	marshalizer              marshal.Marshalizer
	scrPool                  dataRetriever.ShardedDataCacherNotifier
	scrStorer                storage.Storer
	addrConverter            state.AddressConverter
	hasher                   hashing.Hasher
	shardCoordinator         sharding.Coordinator
	broadcastCallbackHandler func(buffToSend []byte)
}

// NewScrInterceptor hooks a new interceptor for transactions
func NewScrInterceptor(
	marshalizer marshal.Marshalizer,
	scrPool dataRetriever.ShardedDataCacherNotifier,
	scrStorer storage.Storer,
	addrConverter state.AddressConverter,
	hasher hashing.Hasher,
	shardCoordinator sharding.Coordinator,
) (*ScrInterceptor, error) {

	if marshalizer == nil {
		return nil, process.ErrNilMarshalizer
	}
	if scrPool == nil {
		return nil, process.ErrNilScrDataPool
	}
	if scrStorer == nil {
		return nil, process.ErrNilScrStorage
	}
	if addrConverter == nil {
		return nil, process.ErrNilAddressConverter
	}
	if hasher == nil {
		return nil, process.ErrNilHasher
	}
	if shardCoordinator == nil {
		return nil, process.ErrNilShardCoordinator
	}

	scrIntercept := &ScrInterceptor{
		marshalizer:      marshalizer,
		scrPool:          scrPool,
		scrStorer:        scrStorer,
		hasher:           hasher,
		addrConverter:    addrConverter,
		shardCoordinator: shardCoordinator,
	}

	return scrIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (scri *ScrInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
	if message == nil {
		return process.ErrNilMessage
	}

	if message.Data() == nil {
		return process.ErrNilDataToProcess
	}

	scrsBuff := make([][]byte, 0)
	err := scri.marshalizer.Unmarshal(&scrsBuff, message.Data())
	if err != nil {
		return err
	}
	if len(scrsBuff) == 0 {
		return process.ErrNoSmartContractResultInMessage
	}

	filteredScrsBuffs := make([][]byte, 0)
	lastErrEncountered := error(nil)
	for _, scrBuff := range scrsBuff {
		scrIntercepted, err := NewInterceptedSmartContractResult(
			scrBuff,
			scri.marshalizer,
			scri.hasher,
			scri.addrConverter,
			scri.shardCoordinator)

		if err != nil {
			lastErrEncountered = err
			continue
		}

		//scr is validated, add it to filtered out scrs
		filteredScrsBuffs = append(filteredScrsBuffs, scrBuff)
		if scrIntercepted.IsAddressedToOtherShards() {
			log.Debug("intercepted scr is for other shards")

			continue
		}

		go scri.processSmartContractResult(scrIntercepted)
	}

	var buffToSend []byte
	filteredOutScrsNeedToBeSend := len(filteredScrsBuffs) > 0 && lastErrEncountered != nil
	if filteredOutScrsNeedToBeSend {
		buffToSend, err = scri.marshalizer.Marshal(filteredScrsBuffs)
		if err != nil {
			return err
		}
	}

	if scri.broadcastCallbackHandler != nil {
		scri.broadcastCallbackHandler(buffToSend)
	}

	return lastErrEncountered
}

// SetBroadcastCallback sets the callback method to send filtered out message
func (scri *ScrInterceptor) SetBroadcastCallback(callback func(buffToSend []byte)) {
	scri.broadcastCallbackHandler = callback
}

func (scri *ScrInterceptor) processSmartContractResult(scr *InterceptedSmartContractResult) {
	//TODO should remove this as it is expensive
	err := scri.scrStorer.Has(scr.Hash())
	isScrInStorage := err == nil
	if isScrInStorage {
		log.Debug("intercepted scr already processed")
		return
	}

	cacherIdentifier := process.ShardCacherIdentifier(scr.SndShard(), scr.RcvShard())
	scri.scrPool.AddData(
		scr.Hash(),
		scr.SmartContractResult(),
		cacherIdentifier,
	)
}
