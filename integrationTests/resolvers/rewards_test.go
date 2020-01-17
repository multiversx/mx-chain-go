package resolvers

import (
	"bytes"
	"testing"

	"github.com/ElrondNetwork/elrond-go/process/factory"
)

func TestRequestResolveRewardsByHashRequestingShardResolvingSameShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardId := uint32(0)
	nResolver, nRequester := createResolverRequester(shardId, shardId)
	headerNonce := uint64(0)
	reward, hash := createReward(headerNonce, shardId)

	//add reward with round 0 in pool
	nResolver.ShardDataPool.RewardTransactions().AddData(hash, reward, "cache")

	//setup header received event
	nRequester.ShardDataPool.RewardTransactions().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received reward tx", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.IntraShardResolver(factory.RewardsTransactionTopic)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash, 0)
	log.LogIfError(err)

	rm.waitWithTimeout()
}

func TestRequestResolveRewardsByHashRequestingShardResolvingOtherShard(t *testing.T) {
	if testing.Short() {
		t.Skip("this is not a short test")
	}

	rm := newReceiverMonitor(t)
	shardIdResolver := uint32(0)
	shardIdRequester := uint32(1)
	nResolver, nRequester := createResolverRequester(shardIdResolver, shardIdRequester)
	headerNonce := uint64(0)
	reward, hash := createReward(headerNonce, shardIdRequester)

	//add reward with round 0 in pool
	nResolver.ShardDataPool.RewardTransactions().AddData(hash, reward, "cache")

	//setup header received event
	nRequester.ShardDataPool.RewardTransactions().RegisterHandler(
		func(key []byte) {
			if bytes.Equal(key, hash) {
				log.Info("received reward tx", "hash", key)
				rm.done()
			}
		},
	)

	//request by hash should work
	resolver, err := nRequester.ResolverFinder.CrossShardResolver(factory.RewardsTransactionTopic, shardIdResolver)
	log.LogIfError(err)
	err = resolver.RequestDataFromHash(hash, 0)
	log.LogIfError(err)

	rm.waitWithTimeout()
}
