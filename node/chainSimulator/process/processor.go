package process

import (
	"github.com/multiversx/mx-chain-core-go/core"
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
	processComponents := creator.nodeHandler.GetProcessComponents()
	cryptoComponents := creator.nodeHandler.GetCryptoComponents()
	coreComponents := creator.nodeHandler.GetCoreComponents()
	bp := processComponents.BlockProcessor()

	nonce, _, prevHash, prevRandSeed, epoch, prevHeader := creator.getPreviousHeaderData()
	round := coreComponents.RoundHandler().Index()
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

	headerCreationTime := coreComponents.RoundHandler().TimeStamp()
	err = newHeader.SetTimeStamp(uint64(headerCreationTime.Unix()))
	if err != nil {
		return err
	}

	leader, validators, err := processComponents.NodesCoordinator().ComputeConsensusGroup(prevRandSeed, newHeader.GetRound(), shardID, epoch)
	if err != nil {
		return err
	}

	pubKeyBitmap := GeneratePubKeyBitmap(len(validators))
	for idx, validator := range validators {
		isManaged := cryptoComponents.KeysHandler().IsKeyManagedByCurrentNode(validator.PubKey())
		if isManaged {
			continue
		}

		err = UnsetBitInBitmap(idx, pubKeyBitmap)
		if err != nil {
			return err
		}
	}

	err = newHeader.SetPubKeysBitmap(pubKeyBitmap)
	if err != nil {
		return err
	}

	isManaged := cryptoComponents.KeysHandler().IsKeyManagedByCurrentNode(leader.PubKey())
	if !isManaged {
		log.Debug("cannot propose block - leader bls key is missing",
			"leader key", leader.PubKey(),
			"shard", creator.nodeHandler.GetShardCoordinator().SelfId())
		return nil
	}

	signingHandler := cryptoComponents.ConsensusSigningHandler()
	randSeed, err := signingHandler.CreateSignatureForPublicKey(newHeader.GetPrevRandSeed(), leader.PubKey())
	if err != nil {
		return err
	}
	err = newHeader.SetRandSeed(randSeed)
	if err != nil {
		return err
	}

	enableEpochHandler := coreComponents.EnableEpochsHandler()

	header, block, err := bp.CreateBlock(newHeader, func() bool {
		return true
	})
	if err != nil {
		return err
	}

	prevHeaderStartOfEpoch := false
	if prevHeader != nil {
		prevHeaderStartOfEpoch = prevHeader.IsStartOfEpochBlock()
	}
	if prevHeaderStartOfEpoch {
		creator.updatePeerShardMapper(header.GetEpoch())
	}

	headerProof, err := creator.ApplySignaturesAndGetProof(header, prevHeader, enableEpochHandler, validators, leader, pubKeyBitmap)
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

	messenger := creator.nodeHandler.GetBroadcastMessenger()
	err = messenger.BroadcastHeader(header, leader.PubKey())
	if err != nil {
		return err
	}

	if !check.IfNil(headerProof) {
		err = messenger.BroadcastEquivalentProof(headerProof, leader.PubKey())
		if err != nil {
			return err
		}
	}

	err = messenger.BroadcastMiniBlocks(miniBlocks, leader.PubKey())
	if err != nil {
		return err
	}

	return messenger.BroadcastTransactions(transactions, leader.PubKey())
}

func (creator *blocksCreator) updatePeerShardMapper(
	epoch uint32,
) {
	peerShardMapper := creator.nodeHandler.GetProcessComponents().PeerShardMapper()

	nc := creator.nodeHandler.GetProcessComponents().NodesCoordinator()

	eligibleMaps, err := nc.GetAllEligibleValidatorsPublicKeys(epoch)
	if err != nil {
		log.Error("failed to get eligible validators map", "error", err)
		return
	}

	for shardID, eligibleMap := range eligibleMaps {
		for _, pubKey := range eligibleMap {
			peerID := creator.nodeHandler.GetBasePeers()[shardID]

			log.Debug("added custom peer mapping", "peerID", peerID.Pretty(), "shardID", shardID, "addrs", pubKey)
			peerShardMapper.UpdatePeerIDInfo(peerID, pubKey, shardID)
		}
	}

}

// ApplySignaturesAndGetProof -
func (creator *blocksCreator) ApplySignaturesAndGetProof(
	header data.HeaderHandler,
	prevHeader data.HeaderHandler,
	enableEpochHandler common.EnableEpochsHandler,
	validators []nodesCoordinator.Validator,
	leader nodesCoordinator.Validator,
	pubKeyBitmap []byte,
) (*dataBlock.HeaderProof, error) {
	nilPrevHeader := check.IfNil(prevHeader)

	err := creator.setHeaderSignatures(header, leader.PubKey(), validators)
	if err != nil {
		return nil, err
	}

	coreComponents := creator.nodeHandler.GetCoreComponents()
	hasher := coreComponents.Hasher()
	marshaller := coreComponents.InternalMarshalizer()
	headerHash, err := core.CalculateHash(marshaller, hasher, header)
	if err != nil {
		return nil, err
	}

	pubKeys := extractValidatorPubKeys(validators)
	newHeaderSig, err := creator.generateAggregatedSignature(headerHash, header.GetEpoch(), header.GetPubKeysBitmap(), pubKeys)
	if err != nil {
		return nil, err
	}

	var headerProof *dataBlock.HeaderProof
	shouldAddCurrentProof := !nilPrevHeader && enableEpochHandler.IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch())
	if shouldAddCurrentProof {
		headerProof = createProofForHeader(pubKeyBitmap, newHeaderSig, headerHash, header)
		creator.nodeHandler.GetDataComponents().Datapool().Headers().AddHeader(headerHash, header)
		err = creator.nodeHandler.GetProcessComponents().HeaderSigVerifier().VerifyHeaderProof(headerProof)
		if err != nil {
			return nil, err
		}

		dataPool := creator.nodeHandler.GetDataComponents().Datapool()
		_ = dataPool.Proofs().AddProof(headerProof)
	}

	return headerProof, nil
}

func createProofForHeader(pubKeyBitmap, signature, headerHash []byte, header data.HeaderHandler) *dataBlock.HeaderProof {
	return &dataBlock.HeaderProof{
		PubKeysBitmap:       pubKeyBitmap,
		AggregatedSignature: signature,
		HeaderHash:          headerHash,
		HeaderEpoch:         header.GetEpoch(),
		HeaderNonce:         header.GetNonce(),
		HeaderShardId:       header.GetShardID(),
		HeaderRound:         header.GetRound(),
		IsStartOfEpoch:      header.IsStartOfEpochBlock(),
	}
}

func (creator *blocksCreator) getPreviousHeaderData() (nonce, round uint64, prevHash, prevRandSeed []byte, epoch uint32, currentHeader data.HeaderHandler) {
	chainHandler := creator.nodeHandler.GetChainHandler()
	currentHeader = chainHandler.GetCurrentBlockHeader()

	if currentHeader != nil {
		nonce, round = currentHeader.GetNonce(), currentHeader.GetRound()
		prevHash = chainHandler.GetCurrentBlockHeaderHash()
		prevRandSeed = currentHeader.GetRandSeed()
		epoch = currentHeader.GetEpoch()
		return
	}

	roundHandler := creator.nodeHandler.GetCoreComponents().RoundHandler()
	prevHash = chainHandler.GetGenesisHeaderHash()
	prevRandSeed = chainHandler.GetGenesisHeader().GetRandSeed()
	round = uint64(roundHandler.Index()) - 1
	epoch = chainHandler.GetGenesisHeader().GetEpoch()
	nonce = chainHandler.GetGenesisHeader().GetNonce()

	return
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

	isEquivalentMessageEnabled := creator.nodeHandler.GetCoreComponents().EnableEpochsHandler().IsFlagEnabledInEpoch(common.AndromedaFlag, header.GetEpoch())
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

	totalKey := 0
	for idx, pubKey := range pubKeys {
		isManaged := creator.nodeHandler.GetCryptoComponents().KeysHandler().IsKeyManagedByCurrentNode([]byte(pubKey))
		if !isManaged {

			continue
		}

		totalKey++
		if _, err = signingHandler.CreateSignatureShareForPublicKey(headerHash, uint16(idx), epoch, []byte(pubKey)); err != nil {
			return nil, err
		}
	}

	aggSig, err := signingHandler.AggregateSigs(pubKeysBitmap, epoch)
	if err != nil {
		log.Warn("total", "total", totalKey, "err", err)
		return nil, err
	}

	return aggSig, nil
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

// UnsetBitInBitmap will unset a bit from provided bit based on the provided index
func UnsetBitInBitmap(index int, bitmap []byte) error {
	if index/8 >= len(bitmap) {
		return common.ErrWrongSizeBitmap
	}
	bitmap[index/8] = bitmap[index/8] & ^(1 << uint8(index%8))

	return nil
}
