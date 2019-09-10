package rewardTransaction

import (
    "github.com/ElrondNetwork/elrond-go/core/logger"
    "github.com/ElrondNetwork/elrond-go/data/state"
    "github.com/ElrondNetwork/elrond-go/dataRetriever"
    "github.com/ElrondNetwork/elrond-go/hashing"
    "github.com/ElrondNetwork/elrond-go/marshal"
    "github.com/ElrondNetwork/elrond-go/p2p"
    "github.com/ElrondNetwork/elrond-go/process"
    "github.com/ElrondNetwork/elrond-go/sharding"
    "github.com/ElrondNetwork/elrond-go/storage"
)

var log = logger.DefaultLogger()

// RewardTxInterceptor is used for intercepting reward transactions and storing them into a datapool
type RewardTxInterceptor struct {
    marshalizer              marshal.Marshalizer
    rewardTxPool             dataRetriever.ShardedDataCacherNotifier
    rewardTxStorer           storage.Storer
    addrConverter            state.AddressConverter
    hasher                   hashing.Hasher
    shardCoordinator         sharding.Coordinator
    broadcastCallbackHandler func(buffToSend []byte)
}

// NewRewardTxInterceptor hooks a new interceptor for reward transactions
func NewRewardTxInterceptor(
    marshalizer marshal.Marshalizer,
    rewardTxPool dataRetriever.ShardedDataCacherNotifier,
    rewardTxStorer storage.Storer,
    addrConverter state.AddressConverter,
    hasher hashing.Hasher,
    shardCoordinator sharding.Coordinator,
) (*RewardTxInterceptor, error) {

    if marshalizer == nil {
        return nil, process.ErrNilMarshalizer
    }
    if rewardTxPool == nil {
        return nil, process.ErrNilRewardTxDataPool
    }
    if rewardTxStorer == nil {
        return nil, process.ErrNilRewardsTxStorage
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

    rewardTxIntercept := &RewardTxInterceptor{
        marshalizer:      marshalizer,
        rewardTxPool:     rewardTxPool,
        rewardTxStorer:   rewardTxStorer,
        hasher:           hasher,
        addrConverter:    addrConverter,
        shardCoordinator: shardCoordinator,
    }

    return rewardTxIntercept, nil
}

// ProcessReceivedMessage will be the callback func from the p2p.Messenger and will be called each time a new message was received
// (for the topic this validator was registered to)
func (rti *RewardTxInterceptor) ProcessReceivedMessage(message p2p.MessageP2P) error {
    if message == nil {
        return process.ErrNilMessage
    }

    if message.Data() == nil {
        return process.ErrNilDataToProcess
    }

    rewardTxsBuff := make([][]byte, 0)
    err := rti.marshalizer.Unmarshal(&rewardTxsBuff, message.Data())
    if err != nil {
        return err
    }
    if len(rewardTxsBuff) == 0 {
        return process.ErrNoRewardTransactionInMessage
    }

    filteredRTxBuffs := make([][]byte, 0)
    lastErrEncountered := error(nil)
    for _, rewardTxBuff := range rewardTxsBuff {
        rewardTxIntercepted, err := NewInterceptedRewardTransaction(
            rewardTxBuff,
            rti.marshalizer,
            rti.hasher,
            rti.addrConverter,
            rti.shardCoordinator)

        if err != nil {
            lastErrEncountered = err
            continue
        }

        //reward tx is validated, add it to filtered out reward txs
        filteredRTxBuffs = append(filteredRTxBuffs, rewardTxBuff)
        if rewardTxIntercepted.IsAddressedToOtherShards() {
            log.Debug("intercepted reward transaction is for other shards")

            continue
        }

        go rti.processRewardTransaction(rewardTxIntercepted)
    }

    var buffToSend []byte
    filteredOutRTxsNeedToBeSend := len(filteredRTxBuffs) > 0 && lastErrEncountered != nil
    if filteredOutRTxsNeedToBeSend {
        buffToSend, err = rti.marshalizer.Marshal(filteredRTxBuffs)
        if err != nil {
            return err
        }
    }

    if rti.broadcastCallbackHandler != nil {
        rti.broadcastCallbackHandler(buffToSend)
    }

    return lastErrEncountered
}

// SetBroadcastCallback sets the callback method to send filtered out message
func (rti *RewardTxInterceptor) SetBroadcastCallback(callback func(buffToSend []byte)) {
    rti.broadcastCallbackHandler = callback
}

func (rti *RewardTxInterceptor) processRewardTransaction(rTx *InterceptedRewardTransaction) {
    cacherIdentifier := process.ShardCacherIdentifier(rTx.SndShard(), rTx.RcvShard())
    rti.rewardTxPool.AddData(
        rTx.Hash(),
        rTx.RewardTransaction(),
        cacherIdentifier,
    )
}

// IsInterfaceNil returns true if there is no value under the interface
func (rti *RewardTxInterceptor) IsInterfaceNil() bool {
    if rti == nil {
        return true
    }
    return false
}
