package sposFactory

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/hashing"
	"github.com/multiversx/mx-chain-core-go/marshal"
	"github.com/multiversx/mx-chain-crypto-go"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/consensus/spos/bls"
	"github.com/multiversx/mx-chain-go/outport"
	"github.com/multiversx/mx-chain-go/process"
	"github.com/multiversx/mx-chain-go/sharding"
)

// GetSubroundsFactory returns a subrounds factory depending on the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
	appStatusHandler core.AppStatusHandler,
	outportHandler outport.OutportHandler,
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

// GetConsensusCoreFactory returns a consensus service depending on the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	default:
		return nil, ErrInvalidConsensusType
	}
}

// GetBroadcastMessenger returns a consensus service depending on the given parameter
func GetBroadcastMessenger(
	marshalizer marshal.Marshalizer,
	hasher hashing.Hasher,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	peerSignatureHandler crypto.PeerSignatureHandler,
	headersSubscriber consensus.HeadersPoolSubscriber,
	interceptorsContainer process.InterceptorsContainer,
	alarmScheduler core.TimersScheduler,
	keysHandler consensus.KeysHandler,
) (consensus.BroadcastMessenger, error) {

	if check.IfNil(shardCoordinator) {
		return nil, spos.ErrNilShardCoordinator
	}

	commonMessengerArgs := broadcast.CommonMessengerArgs{
		Marshalizer:                marshalizer,
		Hasher:                     hasher,
		Messenger:                  messenger,
		ShardCoordinator:           shardCoordinator,
		PeerSignatureHandler:       peerSignatureHandler,
		HeadersSubscriber:          headersSubscriber,
		MaxDelayCacheSize:          maxDelayCacheSize,
		MaxValidatorDelayCacheSize: maxDelayCacheSize,
		InterceptorsContainer:      interceptorsContainer,
		AlarmScheduler:             alarmScheduler,
		KeysHandler:                keysHandler,
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
