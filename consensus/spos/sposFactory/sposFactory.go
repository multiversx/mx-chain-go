package sposFactory

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go/consensus/broadcast/delayed"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/consensus/spos/bls"
)

// GetSubroundsFactory returns a subrounds factory depending of the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
	appStatusHandler core.AppStatusHandler,
	outportHandler consensus.OutportHandler,
	chainID []byte,
	currentPid core.PeerID,
) (spos.SubroundsFactory, error) {
	switch consensusType {
	case blsConsensusType:
		subRoundFactoryBls, err := bls.NewSubroundsFactory(
			consensusDataContainer,
			consensusState,
			worker,
			chainID,
			currentPid,
			appStatusHandler,
		)
		if err != nil {
			return nil, err
		}

		subRoundFactoryBls.SetOutportHandler(outportHandler)

		return subRoundFactoryBls, nil
	default:
		return nil, ErrInvalidConsensusType
	}
}

// GetConsensusCoreFactory returns a consensus service depending of the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	default:
		return nil, ErrInvalidConsensusType
	}
}

// GetBroadcastMessenger returns a consensus service depending of the given parameter
func GetBroadcastMessenger(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	messenger consensus.P2PMessenger,
	shardCoordinator consensus.ShardCoordinator,
	privateKey crypto.PrivateKey,
	peerSignatureHandler crypto.PeerSignatureHandler,
	headersSubscriber consensus.HeadersPoolSubscriber,
	interceptorsContainer consensus.InterceptorsContainer,
	alarmScheduler core.TimersScheduler,
	headersCache consensus.Cacher,
) (consensus.BroadcastMessenger, error) {

	if check.IfNil(shardCoordinator) {
		return nil, spos.ErrNilShardCoordinator
	}

	dbbArgs := &delayed.ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: interceptorsContainer,
		HeadersSubscriber:     headersSubscriber,
		LeaderCacheSize:       maxDelayCacheSize,
		ValidatorCacheSize:    maxDelayCacheSize,
		ShardCoordinator:      shardCoordinator,
		AlarmScheduler:        alarmScheduler,
		HeadersCache:          headersCache,
	}

	dbb, err := delayed.NewDelayedBlockBroadcaster(dbbArgs)
	if err != nil {
		return nil, err
	}

	commonMessengerArgs := broadcast.CommonMessengerArgs{
		Marshalizer:                marshalizer,
		Hasher:                     hasher,
		Messenger:                  messenger,
		PrivateKey:                 privateKey,
		ShardCoordinator:           shardCoordinator,
		PeerSignatureHandler:       peerSignatureHandler,
		HeadersSubscriber:          headersSubscriber,
		MaxDelayCacheSize:          maxDelayCacheSize,
		MaxValidatorDelayCacheSize: maxDelayCacheSize,
		InterceptorsContainer:      interceptorsContainer,
		AlarmScheduler:             alarmScheduler,
		DelayBlockBroadcaster:      dbb,
	}

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		shardMessengerArgs := broadcast.ShardChainMessengerArgs{
			CommonMessengerArgs: commonMessengerArgs,
		}

		return broadcast.NewShardChainMessenger(shardMessengerArgs)
	}

	if shardCoordinator.SelfId() == core.MetachainShardId {
		metaMessengerArgs := broadcast.MetaChainMessengerArgs{
			CommonMessengerArgs: commonMessengerArgs,
		}

		return broadcast.NewMetaChainMessenger(metaMessengerArgs)
	}

	return nil, ErrInvalidShardId
}
