package broadcast

import (
	"time"

	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/broadcast/shared"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/process/factory"
)

const validatorDelayPerOrder = time.Second

var _ consensus.BroadcastMessenger = (*shardChainMessenger)(nil)

type shardChainMessenger struct {
	*commonMessenger
}

// ShardChainMessengerArgs holds the arguments for creating a shardChainMessenger instance
type ShardChainMessengerArgs struct {
	CommonMessengerArgs
}

// NewShardChainMessenger creates a new shardChainMessenger object
func NewShardChainMessenger(
	args ShardChainMessengerArgs,
) (*shardChainMessenger, error) {

	err := checkShardChainNilParameters(args)
	if err != nil {
		return nil, err
	}

	cm := &commonMessenger{
		marshalizer:             args.Marshalizer,
		hasher:                  args.Hasher,
		messenger:               args.Messenger,
		shardCoordinator:        args.ShardCoordinator,
		peerSignatureHandler:    args.PeerSignatureHandler,
		keysHandler:             args.KeysHandler,
		delayedBlockBroadcaster: args.DelayedBroadcaster,
	}

	scm := &shardChainMessenger{
		commonMessenger: cm,
	}

	err = scm.delayedBlockBroadcaster.SetBroadcastHandlers(
		scm.BroadcastMiniBlocks,
		scm.BroadcastTransactions,
		scm.BroadcastHeader,
		scm.BroadcastConsensusMessage)
	if err != nil {
		return nil, err
	}

	return scm, nil
}

func checkShardChainNilParameters(
	args ShardChainMessengerArgs,
) error {
	err := checkCommonMessengerNilParameters(args.CommonMessengerArgs)
	if err != nil {
		return err
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

	scm.messenger.Broadcast(factory.ShardBlocksTopic+headerIdentifier, msgHeader)
	scm.messenger.Broadcast(factory.MiniBlocksTopic+selfIdentifier, msgBlockBody)

	return nil
}

// BroadcastHeader will send on in-shard headers topic the header
func (scm *shardChainMessenger) BroadcastHeader(header data.HeaderHandler, pkBytes []byte) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}

	msgHeader, err := scm.marshalizer.Marshal(header)
	if err != nil {
		return err
	}

	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	scm.broadcast(factory.ShardBlocksTopic+shardIdentifier, msgHeader, pkBytes)

	return nil
}

// BroadcastEquivalentProof will broadcast the proof for a header on the shard metachain common topic
func (scm *shardChainMessenger) BroadcastEquivalentProof(proof *block.HeaderProof, pkBytes []byte) error {
	shardIdentifier := scm.shardCoordinator.CommunicationIdentifier(core.MetachainShardId)
	topic := common.EquivalentProofsTopic + shardIdentifier

	return scm.broadcastEquivalentProof(proof, pkBytes, topic)
}

// BroadcastBlockDataLeader broadcasts the block data as consensus group leader
func (scm *shardChainMessenger) BroadcastBlockDataLeader(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	pkBytes []byte,
) error {
	if check.IfNil(header) {
		return spos.ErrNilHeader
	}
	if len(miniBlocks) == 0 {
		return nil
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		return err
	}

	metaMiniBlocks, metaTransactions := scm.extractMetaMiniBlocksAndTransactions(miniBlocks, transactions)

	broadcastData := &shared.DelayedBroadcastData{
		HeaderHash:     headerHash,
		MiniBlocksData: miniBlocks,
		Transactions:   transactions,
		PkBytes:        pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetLeaderData(broadcastData)
	if err != nil {
		return err
	}

	go scm.BroadcastBlockData(metaMiniBlocks, metaTransactions, pkBytes, common.ExtraDelayForBroadcastBlockInfo)
	return nil
}

// PrepareBroadcastHeaderValidator prepares the validator header broadcast in case leader broadcast fails
func (scm *shardChainMessenger) PrepareBroadcastHeaderValidator(
	header data.HeaderHandler,
	_ map[uint32][]byte,
	_ map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if check.IfNil(header) {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", spos.ErrNilHeader)
		return
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}

	vData := &shared.ValidatorHeaderBroadcastData{
		HeaderHash: headerHash,
		Header:     header,
		Order:      uint32(idx),
		PkBytes:    pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetHeaderForValidator(vData)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastHeaderValidator", "error", err)
		return
	}
}

// PrepareBroadcastBlockDataValidator prepares the validator block data broadcast in case leader broadcast fails
func (scm *shardChainMessenger) PrepareBroadcastBlockDataValidator(
	header data.HeaderHandler,
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
	idx int,
	pkBytes []byte,
) {
	if check.IfNil(header) {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", spos.ErrNilHeader)
		return
	}
	if len(miniBlocks) == 0 {
		return
	}

	headerHash, err := core.CalculateHash(scm.marshalizer, scm.hasher, header)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", err)
		return
	}

	broadcastData := &shared.DelayedBroadcastData{
		HeaderHash:     headerHash,
		Header:         header,
		MiniBlocksData: miniBlocks,
		Transactions:   transactions,
		Order:          uint32(idx),
		PkBytes:        pkBytes,
	}

	err = scm.delayedBlockBroadcaster.SetValidatorData(broadcastData)
	if err != nil {
		log.Error("shardChainMessenger.PrepareBroadcastBlockDataValidator", "error", err)
		return
	}
}

// Close closes all the started infinite looping goroutines and subcomponents
func (scm *shardChainMessenger) Close() {
	scm.delayedBlockBroadcaster.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (scm *shardChainMessenger) IsInterfaceNil() bool {
	return scm == nil
}
