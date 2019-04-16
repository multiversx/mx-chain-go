package mock

import (
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos"
	"github.com/ElrondNetwork/elrond-go-sandbox/consensus/spos/bn"
	"github.com/ElrondNetwork/elrond-go-sandbox/crypto"
	"github.com/ElrondNetwork/elrond-go-sandbox/data"
	"github.com/ElrondNetwork/elrond-go-sandbox/data/block"
	"time"
)

func InitChronologyHandlerMock() consensus.ChronologyHandler {
	chr := &ChronologyHandlerMock{}
	return chr
}

func InitBlockProcessorMock() *BlockProcessorMock {
	blockProcessorMock := &BlockProcessorMock{}
	blockProcessorMock.CreateBlockCalled = func(round int32, haveTime func() bool) (data.BodyHandler, error) {
		emptyBlock := make(block.Body, 0)

		return emptyBlock, nil
	}
	blockProcessorMock.CommitBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler) error {
		return nil
	}
	blockProcessorMock.RevertAccountStateCalled = func() {}
	blockProcessorMock.ProcessBlockCalled = func(blockChain data.ChainHandler, header data.HeaderHandler, body data.BodyHandler, haveTime func() time.Duration) error {
		return nil
	}
	blockProcessorMock.GetRootHashCalled = func() []byte {
		return []byte{}
	}
	blockProcessorMock.CreateBlockHeaderCalled = func(body data.BodyHandler) (header data.HeaderHandler, e error) {
		return &block.Header{RootHash: blockProcessorMock.GetRootHashCalled()}, nil
	}
	return blockProcessorMock
}

func InitMultiSignerMock() *BelNevMock {
	multiSigner := NewMultiSigner()
	multiSigner.CreateCommitmentMock = func() ([]byte, []byte) {
		return []byte("commSecret"), []byte("commitment")
	}
	multiSigner.VerifySignatureShareMock = func(index uint16, sig []byte, bitmap []byte) error {
		return nil
	}
	multiSigner.VerifyMock = func(bitmap []byte) error {
		return nil
	}
	multiSigner.AggregateSigsMock = func(bitmap []byte) ([]byte, error) {
		return []byte("aggregatedSig"), nil
	}
	multiSigner.AggregateCommitmentsMock = func(bitmap []byte) error {
		return nil
	}
	multiSigner.CreateSignatureShareMock = func(bitmap []byte) ([]byte, error) {
		return []byte("partialSign"), nil
	}
	return multiSigner
}

func InitKeys() (*KeyGenMock, *PrivateKeyMock, *PublicKeyMock) {
	toByteArrayMock := func() ([]byte, error) {
		return []byte("byteArray"), nil
	}
	privKeyMock := &PrivateKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	pubKeyMock := &PublicKeyMock{
		ToByteArrayMock: toByteArrayMock,
	}
	privKeyFromByteArr := func(b []byte) (crypto.PrivateKey, error) {
		return privKeyMock, nil
	}
	pubKeyFromByteArr := func(b []byte) (crypto.PublicKey, error) {
		return pubKeyMock, nil
	}
	keyGenMock := &KeyGenMock{
		PrivateKeyFromByteArrayMock: privKeyFromByteArr,
		PublicKeyFromByteArrayMock:  pubKeyFromByteArr,
	}
	return keyGenMock, privKeyMock, pubKeyMock
}

func CreateEligibleList(size int) []string {
	eligibleList := make([]string, 0)
	for i := 0; i < size; i++ {
		eligibleList = append(eligibleList, string(i+65))
	}
	return eligibleList
}

func InitConsensusState() *spos.ConsensusState {
	consensusGroupSize := 9
	eligibleList := CreateEligibleList(consensusGroupSize)
	indexLeader := 1
	rcns := spos.NewRoundConsensus(
		eligibleList,
		consensusGroupSize,
		eligibleList[indexLeader])

	rcns.SetConsensusGroup(eligibleList)
	rcns.ResetRoundState()

	PBFTThreshold := consensusGroupSize*2/3 + 1

	rthr := spos.NewRoundThreshold()
	rthr.SetThreshold(bn.SrBlock, 1)
	rthr.SetThreshold(bn.SrCommitmentHash, PBFTThreshold)
	rthr.SetThreshold(bn.SrBitmap, PBFTThreshold)
	rthr.SetThreshold(bn.SrCommitment, PBFTThreshold)
	rthr.SetThreshold(bn.SrSignature, PBFTThreshold)

	rstatus := spos.NewRoundStatus()
	rstatus.ResetRoundStatus()

	cns := spos.NewConsensusState(
		rcns,
		rthr,
		rstatus,
	)

	cns.Data = []byte("X")
	cns.RoundIndex = 0
	return cns
}
