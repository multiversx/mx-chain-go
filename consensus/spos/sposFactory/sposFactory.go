package sposFactory

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/broadcast"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bls"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/marshal"
	"github.com/ElrondNetwork/elrond-go-sandbox/ntp"
	"github.com/ElrondNetwork/elrond-go-sandbox/sharding"
)

// GetSubroundsFactory returns a subrounds factory depending of the given parameter
func GetSubroundsFactory(
	consensusDataContainer spos.ConsensusCoreHandler,
	consensusState *spos.ConsensusState,
	worker spos.WorkerHandler,
	consensusType string,
) (spos.SubroundsFactory, error) {

	switch consensusType {
	case blsConsensusType:
		return bls.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
	case bnConsensusType:
		return bn.NewSubroundsFactory(consensusDataContainer, consensusState, worker)
	}

	return nil, ErrInvalidConsensusType
}

// GetConsensusCoreFactory returns a consensus service depending of the given parameter
func GetConsensusCoreFactory(consensusType string) (spos.ConsensusService, error) {
	switch consensusType {
	case blsConsensusType:
		return bls.NewConsensusService()
	case bnConsensusType:
		return bn.NewConsensusService()
	}

	return nil, ErrInvalidConsensusType
}

// GetBroadcastMessenger returns a consensus service depending of the given parameter
func GetBroadcastMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	singleSigner crypto.SingleSigner,
	syncTimer ntp.SyncTimer,
) (consensus.BroadcastMessenger, error) {

	if shardCoordinator.SelfId() < shardCoordinator.NumberOfShards() {
		return broadcast.NewShardChainMessenger(marshalizer, messenger, privateKey, shardCoordinator, singleSigner, syncTimer)
	}

	if shardCoordinator.SelfId() == sharding.MetachainShardId {
		return broadcast.NewMetaChainMessenger(marshalizer, messenger, privateKey, shardCoordinator, singleSigner)
	}

	return nil, ErrInvalidShardId
}
