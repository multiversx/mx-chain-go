package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type shardChainMessenger struct {
	*commonMessenger
	marshalizer      marshal.Marshalizer
	messenger        consensus.P2PMessenger
	shardCoordinator sharding.Coordinator
}

// NewShardChainMessenger creates a new shardChainMessenger object
func NewShardChainMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*shardChainMessenger, error) {

	err := checkShardChainNilParameters(marshalizer, messenger, shardCoordinator, privateKey, singleSigner)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:      marshalizer,
		messenger:        messenger,
		privateKey:       privateKey,
		shardCoordinator: shardCoordinator,
		singleSigner:     singleSigner,
	}

	scm := &shardChainMessenger{
		commonMessenger:  cm,
		marshalizer:      marshalizer,
		messenger:        messenger,
		shardCoordinator: shardCoordinator,
	}

	return scm, nil
}

func checkShardChainNilParameters(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	shardCoordinator sharding.Coordinator,
	privateKey crypto.PrivateKey,
	singleSigner crypto.SingleSigner,
) error {
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return spos.ErrNilMarshalizer
	}
	if messenger == nil || messenger.IsInterfaceNil() {
		return spos.ErrNilMessenger
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return spos.ErrNilShardCoordinator
	}
	if privateKey == nil || privateKey.IsInterfaceNil() {
		return spos.ErrNilPrivateKey
	}
	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return spos.ErrNilSingleSigner
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *shardChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if blockBody == nil || blockBody.IsInterfaceNil() {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if header == nil || header.IsInterfaceNil() {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	msgBlockBody, err := scm.marshalizer.Marshal(blockBody)
	if err != nil {
		return err
	}

	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	go scm.messenger.Broadcast(factory.HeadersTopic+selfIdentifier, msgHeader)
	go scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastShardHeader will send on shard headers for metachain topic the header
func (scm *shardChainMessenger) BroadcastShardHeader(header data.HeaderHandler) error {
	if header == nil || header.IsInterfaceNil() {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardHeaderForMetachainTopic := factory.ShardHeadersForMetachainTopic +
		scm.shardCoordinator.CommunicationIdentifier(sharding.MetachainShardId)

	go scm.messenger.Broadcast(shardHeaderForMetachainTopic, msgHeader)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if header == nil || header.IsInterfaceNil() {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	go scm.messenger.Broadcast(factory.HeadersTopic+selfIdentifier, msgHeader)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	if scm == nil {
		return true
	}
	return false
}
