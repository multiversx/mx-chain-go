package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

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
	if check.IfNil(marshalizer) {
		return spos.ErrNilMarshalizer
	}
	if check.IfNil(messenger) {
		return spos.ErrNilMessenger
	}
	if check.IfNil(shardCoordinator) {
		return spos.ErrNilShardCoordinator
	}
	if check.IfNil(privateKey) {
		return spos.ErrNilPrivateKey
	}
	if check.IfNil(singleSigner) {
		return spos.ErrNilSingleSigner
	}

	return nil
}

// BroadcastBlock will send on in-shard headers topic and on in-shard miniblocks topic the header and block body
func (scm *shardChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if check.IfNil(blockBody) {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := scm.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	headerIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	selfIdentifier := scm.shardCoordinator.CommunicationIdentifier(scm.shardCoordinator.SelfId())

	go scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, msgHeader)
	go scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	go scm.messenger.Broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
