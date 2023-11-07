package process

import (
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
)

type manualRoundHandler interface {
	IncrementIndex()
}

type blocksCreator struct {
	nodeHandler NodeHandler
	blsKeyBytes []byte
}

// NewBlocksCreator will create a new instance of blocksCreator
func NewBlocksCreator(nodeHandler NodeHandler, blsKeyBytes []byte) (*blocksCreator, error) {
	return &blocksCreator{
		nodeHandler: nodeHandler,
		blsKeyBytes: blsKeyBytes,
	}, nil
}

// IncrementRound will increment the current round
func (creator *blocksCreator) IncrementRound() {
	roundHandler := creator.nodeHandler.GetCoreComponents().RoundHandler()
	manual := roundHandler.(manualRoundHandler)
	manual.IncrementIndex()
}

// CreateNewBlock creates and process a new block
func (creator *blocksCreator) CreateNewBlock() error {
	bp := creator.nodeHandler.GetProcessComponents().BlockProcessor()

	nonce, round, prevHash, prevRandSeed := creator.getPreviousHeaderData()
	newHeader, err := bp.CreateNewHeader(round+1, nonce+1)
	if err != nil {
		return err
	}
	err = newHeader.SetShardID(creator.nodeHandler.GetShardCoordinator().SelfId())
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

	// TODO set the timestamp but refactor the baseForkDetector.computeGenesisTimeFromHeader function
	// err = newHeader.SetTimeStamp(uint64(time.Now().Unix()))
	// if err != nil {
	//	return err
	// }

	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()
	randSeed, err := signingHandler.CreateSignatureForPublicKey(newHeader.GetPrevRandSeed(), creator.blsKeyBytes)
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

	err = creator.setHeaderSignatures(header)
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

	err = creator.nodeHandler.GetBroadcastMessenger().BroadcastHeader(header, creator.blsKeyBytes)
	if err != nil {
		return err
	}

	return creator.nodeHandler.GetBroadcastMessenger().BroadcastBlockDataLeader(header, miniBlocks, transactions, creator.blsKeyBytes)
}

func (creator *blocksCreator) getPreviousHeaderData() (nonce, round uint64, prevHash, prevRandSeed []byte) {
	currentHeader := creator.nodeHandler.GetChainHandler().GetCurrentBlockHeader()

	if currentHeader != nil {
		nonce, round = currentHeader.GetNonce(), currentHeader.GetRound()
		prevHash = creator.nodeHandler.GetChainHandler().GetCurrentBlockHeaderHash()
		prevRandSeed = currentHeader.GetRandSeed()

		return
	}

	prevHash = creator.nodeHandler.GetChainHandler().GetGenesisHeaderHash()
	prevRandSeed = creator.nodeHandler.GetChainHandler().GetGenesisHeader().GetRandSeed()

	return
}

func (creator *blocksCreator) setHeaderSignatures(header data.HeaderHandler) error {
	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()
	headerClone := header.ShallowClone()
	_ = headerClone.SetPubKeysBitmap(nil)

	marshalizedHdr, err := creator.nodeHandler.GetCoreComponents().InternalMarshalizer().Marshal(headerClone)
	if err != nil {
		return err
	}

	err = signingHandler.Reset([]string{string(creator.blsKeyBytes)})
	if err != nil {
		return err
	}

	headerHash := creator.nodeHandler.GetCoreComponents().Hasher().Compute(string(marshalizedHdr))
	_, err = signingHandler.CreateSignatureShareForPublicKey(
		headerHash,
		uint16(0),
		header.GetEpoch(),
		creator.blsKeyBytes,
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

	leaderSignature, err := creator.createLeaderSignature(header)
	if err != nil {
		return err
	}

	err = header.SetLeaderSignature(leaderSignature)
	if err != nil {
		return err
	}

	return nil
}

func (creator *blocksCreator) createLeaderSignature(header data.HeaderHandler) ([]byte, error) {
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

	return signingHandler.CreateSignatureForPublicKey(marshalizedHdr, creator.blsKeyBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (creator *blocksCreator) IsInterfaceNil() bool {
	return creator == nil
}
