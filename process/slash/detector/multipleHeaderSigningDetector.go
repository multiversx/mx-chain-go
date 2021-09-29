package detector

import (
	"errors"

	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/block/interceptedBlocks"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/sharding"
)

// SigningSlashingDetector - checks for slashable events for headers
type SigningSlashingDetector struct {
	slashingCache    detectorCache
	roundHashesCache roundDataCache
	roundHeaderMap   map[string]string
	nodesCoordinator sharding.NodesCoordinator
	roundHandler     process.RoundHandler
	hasher           hashing.Hasher
	marshaller       marshal.Marshalizer
}

// NewSigningSlashingDetector - creates a new header slashing detector for multiple signatures
func NewSigningSlashingDetector() slash.SlashingDetector {
	return &SigningSlashingDetector{}
}

// VerifyData - checks if an intercepted data represents a slashable event
func (ssd *SigningSlashingDetector) VerifyData(data process.InterceptedData) (slash.SlashingProofHandler, error) {
	header, castOk := data.(*interceptedBlocks.InterceptedHeader)
	if !castOk {
		return nil, process.ErrCannotCastInterceptedDataToHeader
	}

	round := header.HeaderHandler().GetRound()
	if !ssd.isRoundRelevant(round) {
		return nil, process.ErrHeaderRoundNotRelevant
	}

	headerCopy := header
	headerCopy.HeaderHandler().SetPubKeysBitmap(nil)
	headerCopy.HeaderHandler().SetSignature(nil)
	headerCopy.HeaderHandler().SetLeaderSignature(nil)

	headerBytes, err := ssd.marshaller.Marshal(headerCopy)
	if err != nil {
		return nil, err
	}

	headerHash := ssd.hasher.Compute(string(headerBytes))

	if ssd.roundHashesCache.contains(round, headerHash) {
		return nil, errors.New("smth")
	}

	for _, currData := range ssd.roundHashesCache.headers(round) {
		badBoys := ssd.baietiiCareAuSemnatMaiMulte(currData.header, header.HeaderHandler())
		if len(badBoys) >= 1 {
			for _, badBoy := range badBoys {
				ssd.slashingCache.add(round, badBoy.PubKey(), header)
			}
		}
	}

	ssd.roundHashesCache.add(round, headerHash, header.HeaderHandler())

	// check another signature with the same round and proposer exists, but a different header exists
	// if yes a slashingDetectorResult is returned with a message and the two signatures
	return nil, nil
}

func (ssd *SigningSlashingDetector) baietiiCareAuSemnatMaiMulte(header1 data.HeaderHandler, header2 data.HeaderHandler) []sharding.Validator {
	return nil
}

func (ssd *SigningSlashingDetector) isRoundRelevant(headerRound uint64) bool {
	currRound := uint64(ssd.roundHandler.Index())
	return absDiff(currRound, headerRound) < MaxDeltaToCurrentRound
}

func (ssd *SigningSlashingDetector) getProposerPubKey(header data.HeaderHandler) ([]byte, error) {
	validators, err := ssd.nodesCoordinator.ComputeConsensusGroup(
		header.GetRandSeed(),
		header.GetRound(),
		header.GetShardID(),
		header.GetEpoch())

	if err != nil {
		return nil, err
	}
	if len(validators) == 0 {
		return nil, process.ErrEmptyConsensusGroup
	}

	return validators[0].PubKey(), nil
}

// GenerateProof - creates the SlashingProofHandler for the DetectorResult to be added to the Tx Data Field
func (ssd *SigningSlashingDetector) ValidateProof(proof slash.SlashingProofHandler) error {
	return nil
}
