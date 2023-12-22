package broadcast

import (
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
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

	dbbArgs := &ArgsDelayedBlockBroadcaster{
		InterceptorsContainer: args.InterceptorsContainer,
		HeadersSubscriber:     args.HeadersSubscriber,
		LeaderCacheSize:       args.MaxDelayCacheSize,
		ValidatorCacheSize:    args.MaxValidatorDelayCacheSize,
		ShardCoordinator:      args.ShardCoordinator,
		AlarmScheduler:        args.AlarmScheduler,
		Config:                args.Config,
	}

	dbb, err := NewDelayedBlockBroadcaster(dbbArgs)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:             args.Marshalizer,
		hasher:                  args.Hasher,
		messenger:               args.Messenger,
		shardCoordinator:        args.ShardCoordinator,
		peerSignatureHandler:    args.PeerSignatureHandler,
		delayedBlockBroadcaster: dbb,
		keysHandler:             args.KeysHandler,
	}

	mcm := &metaChainMessenger{
		commonMessenger: cm,
	}

	err = dbb.SetBroadcastHandlers(mcm.BroadcastMiniBlocks, mcm.BroadcastTransactions, mcm.BroadcastHeader, mcm.BroadcastConsensusMessage)
	if err != nil {
		return nil, err
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

	mcm.messenger.Broadcast(factory.MetachainBlocksTopic, msgHeader)
	mcm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on metachain blocks topic the header
func (mcm *metaChainMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := mcm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	mcm.broadcast(factory.MetachainBlocksTopic, msgHeader, pkBytes)

	return nil
}

// BroadcastBlockDataLeader broadcasts the block data as consensus group leader
func (mcm *metaChainMessenger) BroadcastBlockDataLeader(
	_ data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	pkBytes []byte,
) error {
	go mcm.BroadcastBlockData(miniBlocks, transactions, pkBytes, common.ExtraDelayForBroadcastBlockInfo)
	return nil
}

// PrepareBroadcastHeaderValidator prepares the validator header broadcast in case leader broadcast fails
func (mcm *metaChainMessenger) PrepareBroadcastHeaderValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if check.IfNil(header) {
		log.Error("metaChainMessenger.PrepareBroadcastHeaderValidator", "error", spos.ErrNilHeader)
		return
	}

	headerHash, err := core.CalculateHash(mcm.marshalizer, mcm.hasher, header)
	if err != nil {
		log.Error("metaChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}

	vData := &validatorHeaderBroadcastData{
		headerHash:           headerHash,
		header:               header,
		metaMiniBlocksData:   miniBlocks,
		metaTransactionsData: transactions,
		order:                uint32(idx),
		pkBytes:              pkBytes,
	}

	err = mcm.delayedBlockBroadcaster.SetHeaderForValidator(vData)
	if err != nil {
		log.Error("metaChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}
}

// PrepareBroadcastBlockDataValidator prepares the validator fallback broadcast in case leader broadcast fails
func (mcm *metaChainMessenger) PrepareBroadcastBlockDataValidator(
	_ data.HeaderHandler,
	_ map[uint32][]byte,
	_ map[string][][]byte,
	_ int,
	_ []byte,
) {
}

// Close closes all the started infinite looping goroutines and subcomponents
func (mcm *metaChainMessenger) Close() {
	mcm.delayedBlockBroadcaster.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (mcm *metaChainMessenger) IsInterfaceNil() bool {
	return mcm == nil
}
