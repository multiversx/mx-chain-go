package process

import (
	"time"

	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("process-block")

type manualRoundHandler interface {
	IncrementIndex()
}

type blocksCreator struct {
	nodeHandler NodeHandler
}

// NewBlocksCreator will create a new instance of blocksCreator
func NewBlocksCreator(nodeHandler NodeHandler) (*blocksCreator, error) {
	if check.IfNil(nodeHandler) {
		return nil, ErrNilNodeHandler
	}

	return &blocksCreator{
		nodeHandler: nodeHandler,
	}, nil
}

// IncrementRound will increment the current round
func (creator *blocksCreator) IncrementRound() {
	roundHandler := creator.nodeHandler.GetCoreComponents().RoundHandler()
	manual := roundHandler.(manualRoundHandler)
	manual.IncrementIndex()

	creator.nodeHandler.GetStatusCoreComponents().AppStatusHandler().SetUInt64Value(common.MetricCurrentRound, uint64(roundHandler.Index()))
}

// CreateNewBlock creates and process a new block
func (creator *blocksCreator) CreateNewBlock() error {
	bp := creator.nodeHandler.GetProcessComponents().BlockProcessor()

	nonce, _, prevHash, prevRandSeed, epoch := creator.getPreviousHeaderData()
	round := creator.nodeHandler.GetCoreComponents().RoundHandler().Index()
	newHeader, err := bp.CreateNewHeader(uint64(round), nonce+1)
	if err != nil {
		return err
	}

	shardID := creator.nodeHandler.GetShardCoordinator().SelfId()
	err = newHeader.SetShardID(shardID)
	if err != nil {
		return err
	}

	err = newHeader.SetPrevHash(prevHash)
	if err != nil {
		return err
	}

	err = newHeader.SetPrevRandSeed(prevRandSeed)
	if err != nil {
		return err
	}

	err = newHeader.SetPubKeysBitmap([]byte{1})
	if err != nil {
		return err
	}

	err = newHeader.SetChainID([]byte(configs.ChainID))
	if err != nil {
		return err
	}

	headerCreationTime := creator.nodeHandler.GetCoreComponents().RoundHandler().TimeStamp()
	err = newHeader.SetTimeStamp(uint64(headerCreationTime.Unix()))
	if err != nil {
		return err
	}

	validatorsGroup, err := creator.nodeHandler.GetProcessComponents().NodesCoordinator().ComputeConsensusGroup(prevRandSeed, newHeader.GetRound(), shardID, epoch)
	if err != nil {
		return err
	}
	blsKey := validatorsGroup[spos.IndexOfLeaderInConsensusGroup]

	isManaged := creator.nodeHandler.GetCryptoComponents().KeysHandler().IsKeyManagedByCurrentNode(blsKey.PubKey())
	if !isManaged {
		log.Debug("cannot propose block - leader bls key is missing",
			"leader key", blsKey.PubKey(),
			"shard", creator.nodeHandler.GetShardCoordinator().SelfId())
		return nil
	}

	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()
	randSeed, err := signingHandler.CreateSignatureForPublicKey(newHeader.GetPrevRandSeed(), blsKey.PubKey())
	if err != nil {
		return err
	}
	err = newHeader.SetRandSeed(randSeed)
	if err != nil {
		return err
	}

	header, block, err := bp.CreateBlock(newHeader, func() bool {
		return true
	})
	if err != nil {
		return err
	}

	err = creator.setHeaderSignatures(header, blsKey.PubKey())
	if err != nil {
		return err
	}

	header, block, err = bp.ProcessBlock(header, block, func() time.Duration {
		return time.Second
	})
	if err != nil {
		return err
	}

	err = bp.CommitBlock(header, block)
	if err != nil {
		return err
	}

	miniBlocks, transactions, err := bp.MarshalizedDataToBroadcast(header, block)
	if err != nil {
		return err
	}

	err = creator.nodeHandler.GetBroadcastMessenger().BroadcastHeader(header, blsKey.PubKey())
	if err != nil {
		return err
	}

	err = creator.nodeHandler.GetBroadcastMessenger().BroadcastMiniBlocks(miniBlocks, blsKey.PubKey())
	if err != nil {
		return err
	}

	return creator.nodeHandler.GetBroadcastMessenger().BroadcastTransactions(transactions, blsKey.PubKey())
}

func (creator *blocksCreator) getPreviousHeaderData() (nonce, round uint64, prevHash, prevRandSeed []byte, epoch uint32) {
	currentHeader := creator.nodeHandler.GetChainHandler().GetCurrentBlockHeader()

	if currentHeader != nil {
		nonce, round = currentHeader.GetNonce(), currentHeader.GetRound()
		prevHash = creator.nodeHandler.GetChainHandler().GetCurrentBlockHeaderHash()
		prevRandSeed = currentHeader.GetRandSeed()
		epoch = currentHeader.GetEpoch()
		return
	}

	prevHash = creator.nodeHandler.GetChainHandler().GetGenesisHeaderHash()
	prevRandSeed = creator.nodeHandler.GetChainHandler().GetGenesisHeader().GetRandSeed()
	round = uint64(creator.nodeHandler.GetCoreComponents().RoundHandler().Index()) - 1
	epoch = creator.nodeHandler.GetChainHandler().GetGenesisHeader().GetEpoch()
	nonce = creator.nodeHandler.GetChainHandler().GetGenesisHeader().GetNonce()

	return
}

func (creator *blocksCreator) setHeaderSignatures(header data.HeaderHandler, blsKeyBytes []byte) error {
	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()
	headerClone := header.ShallowClone()
	_ = headerClone.SetPubKeysBitmap(nil)

	marshalizedHdr, err := creator.nodeHandler.GetCoreComponents().InternalMarshalizer().Marshal(headerClone)
	if err != nil {
		return err
	}

	err = signingHandler.Reset([]string{string(blsKeyBytes)})
	if err != nil {
		return err
	}

	headerHash := creator.nodeHandler.GetCoreComponents().Hasher().Compute(string(marshalizedHdr))
	_, err = signingHandler.CreateSignatureShareForPublicKey(
		headerHash,
		uint16(0),
		header.GetEpoch(),
		blsKeyBytes,
	)
	if err != nil {
		return err
	}

	sig, err := signingHandler.AggregateSigs(header.GetPubKeysBitmap(), header.GetEpoch())
	if err != nil {
		return err
	}

	err = header.SetSignature(sig)
	if err != nil {
		return err
	}

	leaderSignature, err := creator.createLeaderSignature(header, blsKeyBytes)
	if err != nil {
		return err
	}

	err = header.SetLeaderSignature(leaderSignature)
	if err != nil {
		return err
	}

	return nil
}

func (creator *blocksCreator) createLeaderSignature(header data.HeaderHandler, blsKeyBytes []byte) ([]byte, error) {
	headerClone := header.ShallowClone()
	err := headerClone.SetLeaderSignature(nil)
	if err != nil {
		return nil, err
	}

	marshalizedHdr, err := creator.nodeHandler.GetCoreComponents().InternalMarshalizer().Marshal(headerClone)
	if err != nil {
		return nil, err
	}

	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()

	return signingHandler.CreateSignatureForPublicKey(marshalizedHdr, blsKeyBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *blocksCreator) IsInterfaceNil() bool {
	return creator == nil
}
