package broadcast

import (
	"github.com/ElrondNetwork/elrond-go/consensus"
	"github.com/ElrondNetwork/elrond-go/consensus/spos"
	"github.com/ElrondNetwork/elrond-go/core/alarm"
	"github.com/ElrondNetwork/elrond-go/core/check"
	"github.com/ElrondNetwork/elrond-go/data"
	"github.com/ElrondNetwork/elrond-go/data/block"
	"github.com/ElrondNetwork/elrond-go/process/factory"
)

var _ consensus.BroadcastMessenger = (*metaChainMessenger)(nil)

type metaChainMessenger struct {
	*commonMessenger
}

// MetaChainMessengerArgs holds the arguments for creating a metaChainMessenger instance
type MetaChainMessengerArgs struct {
	CommonMessengerArgs
}

// NewMetaChainMessenger creates a new metaChainMessenger object
func NewMetaChainMessenger(
	args MetaChainMessengerArgs,
) (*metaChainMessenger, error) {

	err := checkMetaChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:           args.Marshalizer,
		messenger:             args.Messenger,
		privateKey:            args.PrivateKey,
		shardCoordinator:      args.ShardCoordinator,
		singleSigner:          args.SingleSigner,
		interceptorsContainer: args.InterceptorsContainer,
		alarm:                 alarm.NewAlarmScheduler(),
	}

	mcm := &metaChainMessenger{
		commonMessenger: cm,
	}

	return mcm, nil
}

func checkMetaChainNilParameters(
	args MetaChainMessengerArgs,
) error {
	return checkCommonMessengerNilParameters(args.CommonMessengerArgs)
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

// SetLeaderDelayBroadcast - not used for metachain nodes
func (mcm *metaChainMessenger) SetLeaderDelayBroadcast(_ []byte, _ map[uint32][]byte, _ map[string][][]byte) error {
	log.Warn("SetLeaderDelayBroadcast not implemented for metachain")
	return nil
}

// SetValidatorDelayBroadcast - not used for metachain nodes
func (mcm *metaChainMessenger) SetValidatorDelayBroadcast(headerHash []byte, prevRandSeed []byte, round uint64, miniBlocks map[uint32][]byte, miniBlockHashes map[uint32]map[string]struct{}, transactions map[string][][]byte, order uint32) error {
	log.Warn("SetValidatorDelayBroadcast not implemented for metachain")
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcm *metaChainMessenger) IsInterfaceNil() bool {
	return mcm == nil
}
