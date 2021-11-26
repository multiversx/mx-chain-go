package detector_test

import (
	"strconv"
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/sliceUtil"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	mock2 "github.com/ElrondNetwork/elrond-go-core/data/mock"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go-crypto/signing/mcl/singlesig"
	"github.com/ElrondNetwork/elrond-go/sharding"
	"github.com/stretchr/testify/require"
)

type sigShareData struct {
	signature []byte
	index     uint32
}

// This function creates a noOfHeaders, which are all signed by a randomly
// selected noOfMaliciousSigners, from allMultiSigners.
// If noOfMaliciousSigners < minRequiredSignatures in a consensus group, then
// the rest of the signers are randomly selected from allMultiSigners
func generateSlashResults(
	b *testing.B,
	hasher hashing.Hasher,
	noOfMaliciousSigners uint32,
	noOfHeaders uint64,
	nodesCoordinator sharding.NodesCoordinator,
	allMultiSigners map[string]multiSignerData,
) map[string]coreSlash.SlashingResult {
	consensusGroupSize := len(allMultiSigners)
	minRequiredSignatures := calcMinRequiredSignatures(consensusGroupSize)
	bitmapSize := calcBitmapSize(consensusGroupSize)
	maliciousSigners := selectMaliciousSigners(noOfMaliciousSigners, allMultiSigners)

	round := uint64(1)
	epoch := uint32(0)
	headersInfo := make([]data.HeaderInfoHandler, 0, noOfHeaders)
	for i := uint64(0); i < noOfHeaders; i++ {
		randomness := []byte(strconv.Itoa(int(i)))
		header := createMetaHeader(i, round, epoch, randomness)
		publicKeys, err := nodesCoordinator.GetConsensusValidatorsPublicKeys(randomness, round, core.MetachainShardId, epoch)
		require.Nil(b, err)

		headerHash, err := core.CalculateHash(marshaller, hasher, header)
		require.Nil(b, err)

		minRequiredSigners, bitmap, sigShares := selectMinRequiredSigners(b,
			headerHash,
			publicKeys,
			allMultiSigners,
			maliciousSigners,
			minRequiredSignatures,
			bitmapSize)

		leaderPubKey := publicKeys[0]
		leaderPrivateKey := allMultiSigners[leaderPubKey].privateKey
		signHeader(b, header, minRequiredSigners, bitmap, sigShares, leaderPrivateKey)

		headerHash, err = core.CalculateHash(marshaller, hasher, header)
		require.Nil(b, err)

		headerInfo := &mock2.HeaderInfoStub{Header: header, Hash: headerHash}
		headersInfo = append(headersInfo, headerInfo)
	}

	slashRes := make(map[string]coreSlash.SlashingResult, noOfMaliciousSigners)
	for maliciousSigner := range maliciousSigners {
		slashRes[maliciousSigner] = coreSlash.SlashingResult{
			Headers:       headersInfo,
			SlashingLevel: calcThreatLevel(noOfHeaders),
		}
	}

	return slashRes
}

func calcMinRequiredSignatures(consensusGroupSize int) uint32 {
	// Min required signatures = 2/3 * consensusGroup + 1
	return uint32(0.67*float32(consensusGroupSize)) + 1
}

func calcBitmapSize(consensusGroupSize int) uint32 {
	bitmapSize := consensusGroupSize / 8
	if consensusGroupSize%8 != 0 {
		bitmapSize++
	}
	return uint32(bitmapSize)
}

func selectMaliciousSigners(noOfMaliciousSigners uint32, allMultiSigData map[string]multiSignerData) map[string]struct{} {
	maliciousSigners := make(map[string]struct{}, noOfMaliciousSigners)

	ct := uint32(0)
	for pubKey := range allMultiSigData {
		if ct > noOfMaliciousSigners {
			break
		}
		maliciousSigners[pubKey] = struct{}{}
		ct++
	}

	return maliciousSigners
}

func createMetaHeader(timeStamp uint64, round uint64, epoch uint32, prevRandSeed []byte) data.HeaderHandler {
	return &block.HeaderV2{
		Header: &block.Header{
			TimeStamp:    timeStamp,
			Round:        round,
			ShardID:      core.MetachainShardId,
			Epoch:        epoch,
			PrevRandSeed: prevRandSeed,
		},
	}
}

func selectMinRequiredSigners(
	b *testing.B,
	headerHash []byte,
	publicKeys []string,
	allMultiSigners map[string]multiSignerData,
	requiredSigners map[string]struct{},
	minRequiredConsensusSize uint32,
	bitmapSize uint32,
) (minRequiredSigners []crypto.MultiSigner, bitmap []byte, sigSharesData []sigShareData) {
	noOfSigners := uint32(0)
	bitmap = make([]byte, bitmapSize)
	sigSharesData = make([]sigShareData, 0, minRequiredConsensusSize)
	minRequiredSigners = make([]crypto.MultiSigner, 0, minRequiredConsensusSize)

	for idx, pubKey := range publicKeys {
		currSigner := allMultiSigners[pubKey].multiSigner
		_, isRequired := requiredSigners[pubKey]

		if idx == leaderGroupIndex || isRequired || noOfSigners < minRequiredConsensusSize {
			err := currSigner.Reset(publicKeys, uint16(idx))
			require.Nil(b, err)

			signature, err := currSigner.CreateSignatureShare(headerHash, []byte(""))
			require.Nil(b, err)

			sliceUtil.SetIndexInBitmap(uint32(idx), bitmap)
			minRequiredSigners = append(minRequiredSigners, currSigner)
			sigSharesData = append(sigSharesData, sigShareData{index: uint32(idx), signature: signature})
		}
		noOfSigners++
	}

	return minRequiredSigners, bitmap, sigSharesData
}

func signHeader(
	b *testing.B,
	header data.HeaderHandler,
	signers []crypto.MultiSigner,
	bitmap []byte,
	sigSharesData []sigShareData,
	leaderPrivateKey crypto.PrivateKey) {
	aggregatedSig, err := calcAggregatedSignature(signers, sigSharesData, bitmap)
	require.Nil(b, err)
	leaderSignature, err := calcLeaderSignature(leaderPrivateKey, header)
	require.Nil(b, err)

	err = header.SetSignature(aggregatedSig)
	require.Nil(b, err)
	err = header.SetLeaderSignature(leaderSignature)
	require.Nil(b, err)
	err = header.SetPubKeysBitmap(bitmap)
	require.Nil(b, err)
}

func calcAggregatedSignature(signers []crypto.MultiSigner, signaturesData []sigShareData, bitmap []byte) ([]byte, error) {
	err := storeSignatureShares(signers, signaturesData)
	if err != nil {
		return nil, err
	}

	return signers[0].AggregateSigs(bitmap)
}

func storeSignatureShares(multiSigners []crypto.MultiSigner, sigsShareData []sigShareData) error {
	signers := len(multiSigners)
	signatures := len(sigsShareData)

	for i := 0; i < signatures; i++ {
		for j := 0; j < signers; j++ {
			signer := multiSigners[j]
			index := uint16(sigsShareData[i].index)
			signature := sigsShareData[i].signature

			err := signer.StoreSignatureShare(index, signature)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func calcLeaderSignature(leaderPrivateKey crypto.PrivateKey, header data.HeaderHandler) ([]byte, error) {
	leader := singlesig.NewBlsSigner()
	headerBytes, err := marshaller.Marshal(header)
	if err != nil {
		return nil, err
	}

	return leader.Sign(leaderPrivateKey, headerBytes)
}

func calcThreatLevel(noOfHeaders uint64) coreSlash.ThreatLevel {
	threatLevel := coreSlash.Zero
	if noOfHeaders == coreSlash.MinSlashableNoOfHeaders {
		threatLevel = coreSlash.Medium
	} else if noOfHeaders >= coreSlash.MinSlashableNoOfHeaders {
		threatLevel = coreSlash.High
	}

	return threatLevel
}
