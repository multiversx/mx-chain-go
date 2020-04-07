package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/crypto"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/marshal"
	"github.com/ElrondNetwork/elrond-go/process/factory"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

type metaChainMessenger struct {
	*commonMessenger
	marshalizer marshal.Marshalizer
	messenger   consensus.P2PMessenger
}

// NewMetaChainMessenger creates a new metaChainMessenger object
func NewMetaChainMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) (*metaChainMessenger, error) {

	err := checkMetaChainNilParameters(marshalizer, messenger, privateKey, shardCoordinator, singleSigner)
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

	mcm := &metaChainMessenger{
		commonMessenger: cm,
		marshalizer:     marshalizer,
		messenger:       messenger,
	}

	return mcm, nil
}

func checkMetaChainNilParameters(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator sharding.Coordinator,
	singleSigner crypto.SingleSigner,
) error {
	if check.IfNil(marshalizer) {
		return spos.ErrNilMarshalizer
	}
	if check.IfNil(messenger) {
		return spos.ErrNilMessenger
	}
	if check.IfNil(privateKey) {
		return spos.ErrNilPrivateKey
	}
	if check.IfNil(shardCoordinator) {
		return spos.ErrNilShardCoordinator
	}
	if check.IfNil(singleSigner) {
		return spos.ErrNilSingleSigner
	}

	return nil
}

// BroadcastBlock will send on metachain blocks topic the header
func (mcm *metaChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if check.IfNil(blockBody) {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if check.IfNil(header) {
		return spos.ErrNilMetaHeader
	}

	msgHeader, err := mcm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	b := blockBody.(*block.Body)
	msgBlockBody, err := mcm.marshalizer.Marshal(b)
	if err != nil {
		return err
	}

	selfIdentifier := mcm.shardCoordinator.CommunicationIdentifier(mcm.shardCoordinator.SelfId())

	go mcm.messenger.Broadcast(factory.MetachainBlocksTopic, msgHeader)
	go mcm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on metachain blocks topic the header
func (mcm *metaChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := mcm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	go mcm.messenger.Broadcast(factory.MetachainBlocksTopic, msgHeader)

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcm *metaChainMessenger) IsInterfaceNil() bool {
	return mcm == nil
}
