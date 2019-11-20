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
	if marshalizer == nil || marshalizer.IsInterfaceNil() {
		return spos.ErrNilMarshalizer
	}
	if messenger == nil || messenger.IsInterfaceNil() {
		return spos.ErrNilMessenger
	}
	if privateKey == nil || privateKey.IsInterfaceNil() {
		return spos.ErrNilPrivateKey
	}
	if shardCoordinator == nil || shardCoordinator.IsInterfaceNil() {
		return spos.ErrNilShardCoordinator
	}
	if singleSigner == nil || singleSigner.IsInterfaceNil() {
		return spos.ErrNilSingleSigner
	}

	return nil
}

// BroadcastBlock will send on metachain blocks topic the header
func (mcm *metaChainMessenger) BroadcastBlock(blockBody data.BodyHandler, header data.HeaderHandler) error {
	if blockBody == nil || blockBody.IsInterfaceNil() {
		return spos.ErrNilBody
	}

	err := blockBody.IntegrityAndValidity()
	if err != nil {
		return err
	}

	if header == nil || header.IsInterfaceNil() {
		return spos.ErrNilMetaHeader
	}

	msgHeader, err := mcm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	msgBlockBody, err := mcm.marshalizer.Marshal(blockBody)
	if err != nil {
		return err
	}

	selfIdentifier := mcm.shardCoordinator.CommunicationIdentifier(mcm.shardCoordinator.SelfId())

	go mcm.messenger.Broadcast(factory.MetachainBlocksTopic, msgHeader)
	go mcm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastShardHeader will send on meta-to-shards topic the header
func (mcm *metaChainMessenger) BroadcastShardHeader(header data.HeaderHandler) error {
	// meta chain does not need to broadcast separately the header, as it have no body and BroadcastBlock does all
	// the job for it, but this method is created to satisfy the BroadcastMessenger interface
	return nil
}

// BroadcastHeader will send on metachain blocks topic the header
func (mcm *metaChainMessenger) BroadcastHeader(header data.HeaderHandler) error {
	if header == nil || header.IsInterfaceNil() {
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
	if mcm == nil {
		return true
	}
	return false
}
