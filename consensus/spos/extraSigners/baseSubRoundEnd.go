package extraSigners

import (
	"fmt"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data"
	"github.com/multiversx/mx-chain-core-go/data/block"

	"github.com/multiversx/mx-chain-go/consensus"
	"github.com/multiversx/mx-chain-go/consensus/spos"
	"github.com/multiversx/mx-chain-go/errors"
)

type baseSubRoundEnd struct {
	signingHandler consensus.SigningHandler
	mbType         block.OutGoingMBType
}

func newBaseSubRoundEnd(
	signingHandler consensus.SigningHandler,
	mbType block.OutGoingMBType,
) (*baseSubRoundEnd, error) {
	if check.IfNil(signingHandler) {
		return nil, spos.ErrNilSigningHandler
	}

	return &baseSubRoundEnd{
		signingHandler: signingHandler,
		mbType:         mbType,
	}, nil
}

func (br *baseSubRoundEnd) verifyAggregatedSignatures(bitmap []byte, header data.HeaderHandler) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil
	}

	return br.signingHandler.Verify(outGoingMb.GetOutGoingOperationsHash(), bitmap, header.GetEpoch())
}

func (br *baseSubRoundEnd) aggregateAndSetSignatures(bitmap []byte, header data.HeaderHandler) ([]byte, error) {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return nil, fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil, nil
	}

	sig, err := br.signingHandler.AggregateSigs(bitmap, header.GetEpoch())
	if err != nil {
		return nil, err
	}

	err = br.signingHandler.SetAggregatedSig(sig)
	if err != nil {
		return nil, err
	}

	return sig, nil
}

func (br *baseSubRoundEnd) setAggregatedSignatureInHeader(header data.HeaderHandler, aggregatedSig []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil
	}

	err := outGoingMb.SetAggregatedSignatureOutGoingOperations(aggregatedSig)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (br *baseSubRoundEnd) signAndSetLeaderSignature(header data.HeaderHandler, leaderPubKey []byte) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetAggregatedSignatureInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil
	}

	leaderMsgToSign := append(
		outGoingMb.GetOutGoingOperationsHash(),
		outGoingMb.GetAggregatedSignatureOutGoingOperations()...)

	leaderSig, err := br.signingHandler.CreateSignatureForPublicKey(leaderMsgToSign, leaderPubKey)
	if err != nil {
		return err
	}

	err = outGoingMb.SetLeaderSignatureOutGoingOperations(leaderSig)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (br *baseSubRoundEnd) setConsensusDataInHeader(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil
	}

	extraSigData, found := cnsMsg.ExtraSignatures[br.mbType.String()]
	if !found {
		return fmt.Errorf("%w for type %s", errExtraSigShareDataNotFound, br.mbType.String())
	}

	err := outGoingMb.SetAggregatedSignatureOutGoingOperations(extraSigData.AggregatedSignatureOutGoingTxData)
	if err != nil {
		return err
	}
	err = outGoingMb.SetLeaderSignatureOutGoingOperations(extraSigData.LeaderSignatureOutGoingTxData)
	if err != nil {
		return err
	}

	return sovHeader.SetOutGoingMiniBlockHeaderHandler(outGoingMb)
}

func (br *baseSubRoundEnd) addLeaderAndAggregatedSignatures(header data.HeaderHandler, cnsMsg *consensus.Message) error {
	sovHeader, castOk := header.(data.SovereignChainHeaderHandler)
	if !castOk {
		return fmt.Errorf("%w in sovereignSubRoundEndOutGoingTxData.SetConsensusDataInHeader", errors.ErrWrongTypeAssertion)
	}

	outGoingMb := sovHeader.GetOutGoingMiniBlockHeaderHandler(int32(br.mbType))
	if check.IfNil(outGoingMb) {
		return nil
	}

	if cnsMsg.ExtraSignatures == nil {
		cnsMsg.ExtraSignatures = make(map[string]*consensus.ExtraSignatureData)
	}

	keyStr := br.mbType.String()
	if _, found := cnsMsg.ExtraSignatures[keyStr]; !found {
		cnsMsg.ExtraSignatures[keyStr] = &consensus.ExtraSignatureData{}
	}

	cnsMsg.ExtraSignatures[keyStr].AggregatedSignatureOutGoingTxData = outGoingMb.GetAggregatedSignatureOutGoingOperations()
	cnsMsg.ExtraSignatures[keyStr].LeaderSignatureOutGoingTxData = outGoingMb.GetLeaderSignatureOutGoingOperations()

	log.Debug("sovereignSubRoundEndOutGoingTxData.AddLeaderAndAggregatedSignatures",
		"AggregatedSignatureOutGoingTxData", cnsMsg.ExtraSignatures[keyStr].AggregatedSignatureOutGoingTxData,
		"LeaderSignatureOutGoingTxData", cnsMsg.ExtraSignatures[keyStr].LeaderSignatureOutGoingTxData,
		"type", keyStr)

	return nil
}
