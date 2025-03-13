package process

import (
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	dataBlock "github.com/multiversx/mx-chain-core-go/data/block"
	"github.com/multiversx/mx-chain-go/common"
	"github.com/multiversx/mx-chain-go/node/chainSimulator/configs"
	"github.com/multiversx/mx-chain-go/sharding/nodesCoordinator"
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

	nonce, _, prevHash, prevRandSeed, epoch, currentHeader := creator.getPreviousHeaderData()
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

	err = newHeader.SetChainID([]byte(configs.ChainID))
	if err != nil {
		return err
	}

	headerCreationTime := creator.nodeHandler.GetCoreComponents().RoundHandler().TimeStamp()
	err = newHeader.SetTimeStamp(uint64(headerCreationTime.Unix()))
	if err != nil {
		return err
	}

	leader, validators, err := creator.nodeHandler.GetProcessComponents().NodesCoordinator().ComputeConsensusGroup(prevRandSeed, newHeader.GetRound(), shardID, epoch)
	if err != nil {
		return err
	}
	blsKey := leader

	pubKeyBitmap := GeneratePubKeyBitmap(len(validators))
	err = newHeader.SetPubKeysBitmap(pubKeyBitmap)
	if err != nil {
		return err
	}

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

	notNil := !check.IfNil(currentHeader)
	enableEpochHandler := creator.nodeHandler.GetCoreComponents().EnableEpochsHandler()
	var previousProof *dataBlock.HeaderProof
	if notNil && enableEpochHandler.IsFlagEnabled(common.EquivalentMessagesFlag) {
		sig, errS := creator.generateSignature(prevHash, blsKey.PubKey(), currentHeader)
		if errS != nil {
			return errS
		}
		previousProof = createProofForHeader(pubKeyBitmap, sig, prevHash, currentHeader)
		_ = creator.nodeHandler.GetDataComponents().Datapool().Proofs().AddProof(previousProof)
	}

	header, block, err := bp.CreateBlock(newHeader, func() bool {
		return true
	})
	if err != nil {
		return err
	}

	if notNil && common.ShouldBlockHavePrevProof(header, enableEpochHandler, common.EquivalentMessagesFlag) {
		validators, err = creator.updatePreviousProofAndAddonHeader(prevHash, currentHeader, header, previousProof)
	}
	if err != nil {
		return err
	}

	err = creator.setHeaderSignatures(header, blsKey.PubKey(), validators)
	if err != nil {
		return err
	}

	marshalizedHdr, err := creator.nodeHandler.GetCoreComponents().
		InternalMarshalizer().Marshal(header)
	if err != nil {
		return err
	}
	headerHash := creator.nodeHandler.GetCoreComponents().Hasher().Compute(string(marshalizedHdr))

	pubKeys := extractValidatorPubKeys(validators)
	newHeaderSig, err := creator.generateAggregatedSignature(headerHash, header.GetEpoch(), header.GetPubKeysBitmap(), pubKeys)
	if err != nil {
		return err
	}

	var headerProof *dataBlock.HeaderProof
	shouldAddCurrentProof := notNil && enableEpochHandler.IsFlagEnabled(common.EquivalentMessagesFlag)
	if shouldAddCurrentProof {
		headerProof = createProofForHeader(pubKeyBitmap, newHeaderSig, headerHash, header)
		_ = creator.nodeHandler.GetDataComponents().Datapool().Proofs().AddProof(headerProof)
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

	if shouldAddCurrentProof {
		err = creator.nodeHandler.GetBroadcastMessenger().BroadcastEquivalentProof(headerProof, blsKey.PubKey())
		if err != nil {
			return err
		}
	}

	err = creator.nodeHandler.GetBroadcastMessenger().BroadcastMiniBlocks(miniBlocks, blsKey.PubKey())
	if err != nil {
		return err
	}

	return creator.nodeHandler.GetBroadcastMessenger().BroadcastTransactions(transactions, blsKey.PubKey())
}

func (creator *blocksCreator) updatePreviousProofAndAddonHeader(currentHeaderHash []byte, currentHeader, newHeader data.HeaderHandler, previousProof *dataBlock.HeaderProof) ([]nodesCoordinator.Validator, error) {
	selectionEpoch := currentHeader.GetEpoch()
	if currentHeader.IsStartOfEpochBlock() {
		selectionEpoch = selectionEpoch - 1
	}

	_, validators, err := creator.nodeHandler.GetProcessComponents().NodesCoordinator().ComputeConsensusGroup(currentHeader.GetPrevRandSeed(), currentHeader.GetRound(), currentHeader.GetShardID(), selectionEpoch)
	if err != nil {
		return nil, err
	}

	previousProof.PubKeysBitmap = GeneratePubKeyBitmap(len(validators))
	previousProof.AggregatedSignature, err = creator.generateSignatureForProofs(currentHeaderHash, previousProof, validators)
	if err != nil {
		return nil, err
	}

	newHeader.SetPreviousProof(previousProof)

	return validators, nil
}

func createProofForHeader(pubKeyBitmap, signature, headerHash []byte, header data.HeaderHandler) *dataBlock.HeaderProof {
	return &dataBlock.HeaderProof{
		PubKeysBitmap:       pubKeyBitmap,
		AggregatedSignature: signature,
		HeaderHash:          headerHash,
		HeaderEpoch:         header.GetEpoch(),
		HeaderNonce:         header.GetNonce(),
		HeaderShardId:       header.GetShardID(),
		HeaderRound:         header.GetNonce(),
		IsStartOfEpoch:      header.IsStartOfEpochBlock(),
	}
}

func (creator *blocksCreator) getPreviousHeaderData() (nonce, round uint64, prevHash, prevRandSeed []byte, epoch uint32, currentHeader data.HeaderHandler) {
	currentHeader = creator.nodeHandler.GetChainHandler().GetCurrentBlockHeader()

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

func (creator *blocksCreator) generateSignature(headerHash, blsKeyBytes []byte, header data.HeaderHandler) ([]byte, error) {
	return creator.generateAggregatedSignature(
		headerHash,
		header.GetEpoch(),
		header.GetPubKeysBitmap(),
		[]string{string(blsKeyBytes)},
	)
}

func (creator *blocksCreator) generateSignatureForProofs(headerHash []byte, proof *dataBlock.HeaderProof, validators []nodesCoordinator.Validator,
) ([]byte, error) {
	pubKeys := extractValidatorPubKeys(validators)
	return creator.generateAggregatedSignature(headerHash, proof.GetHeaderEpoch(), proof.GetPubKeysBitmap(), pubKeys)
}

func (creator *blocksCreator) setHeaderSignatures(
	header data.HeaderHandler,
	blsKeyBytes []byte,
	validators []nodesCoordinator.Validator,
) error {
	headerClone := header.ShallowClone()
	_ = headerClone.SetPubKeysBitmap(nil)

	marshalizedHdr, err := creator.nodeHandler.GetCoreComponents().
		InternalMarshalizer().Marshal(headerClone)
	if err != nil {
		return err
	}

	headerHash := creator.nodeHandler.GetCoreComponents().Hasher().Compute(string(marshalizedHdr))
	pubKeys := extractValidatorPubKeys(validators)

	sig, err := creator.generateAggregatedSignature(headerHash, header.GetEpoch(), header.GetPubKeysBitmap(), pubKeys)
	if err != nil {
		return err
	}

	isEquivalentMessageEnabled := creator.nodeHandler.GetCoreComponents().EnableEpochsHandler().IsFlagEnabled(common.EquivalentMessagesFlag)
	if !isEquivalentMessageEnabled {
		if err = header.SetSignature(sig); err != nil {
			return err
		}
	}

	leaderSignature, err := creator.createLeaderSignature(header, blsKeyBytes)
	if err != nil {
		return err
	}

	return header.SetLeaderSignature(leaderSignature)
}

func (creator *blocksCreator) generateAggregatedSignature(headerHash []byte, epoch uint32, pubKeysBitmap []byte, pubKeys []string) ([]byte, error) {
	signingHandler := creator.nodeHandler.GetCryptoComponents().ConsensusSigningHandler()
	err := signingHandler.Reset(pubKeys)
	if err != nil {
		return nil, err
	}

	for idx, pubKey := range pubKeys {
		if _, err = signingHandler.CreateSignatureShareForPublicKey(headerHash, uint16(idx), epoch, []byte(pubKey)); err != nil {
			return nil, err
		}
	}

	return signingHandler.AggregateSigs(pubKeysBitmap, epoch)
}

func extractValidatorPubKeys(validators []nodesCoordinator.Validator) []string {
	pubKeys := make([]string, len(validators))
	for i, validator := range validators {
		pubKeys[i] = string(validator.PubKey())
	}
	return pubKeys
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

// GeneratePubKeyBitmap generates a []byte where the first `numOfOnes` bits are set to 1.
func GeneratePubKeyBitmap(numOfOnes int) []byte {
	if numOfOnes <= 0 {
		return nil // Handle invalid cases
	}

	// calculate how many full bytes are needed
	numBytes := (numOfOnes + 7) / 8 // Equivalent to ceil(numOfOnes / 8)
	result := make([]byte, numBytes)

	// fill in the bytes
	for i := 0; i < numBytes; i++ {
		bitsLeft := numOfOnes - (i * 8)
		if bitsLeft >= 8 {
			result[i] = 0xFF // All 8 bits set to 1 (255 in decimal)
		} else {
			result[i] = byte((1 << bitsLeft) - 1) // Only set the needed bits
		}
	}

	return result
}
