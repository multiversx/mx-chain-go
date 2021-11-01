package notifier

import (
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
)

type proofTxData struct {
	round     uint64
	shardID   uint32
	bytes     []byte
	slashType slash.SlashingType
}

func txDataFromMultipleHeaderProposalProof(proof slash.MultipleProposalProofHandler) (*proofTxData, error) {
	protoProof, err := slash.ToProtoMultipleHeaderProposal(proof)
	if err != nil {
		return nil, err
	}
	proofBytes, err := protoProof.Marshal()
	if err != nil {
		return nil, err
	}
	headers := proof.GetHeaders()
	if len(headers) == 0 {
		return nil, process.ErrNotEnoughHeadersProvided
	}

	return &proofTxData{
		round:     headers[0].Header.GetRound(),
		shardID:   headers[0].Header.GetShardID(),
		bytes:     proofBytes,
		slashType: proof.GetType(),
	}, nil
}

func txDataFromMultipleHeaderSigningProof(proof slash.MultipleSigningProofHandler) (*proofTxData, error) {
	protoProof, err := slash.ToProtoMultipleHeaderSign(proof)
	if err != nil {
		return nil, err
	}
	proofBytes, err := protoProof.Marshal()
	if err != nil {
		return nil, err
	}
	pubKeys := proof.GetPubKeys()
	if len(pubKeys) == 0 {
		return nil, process.ErrNotEnoughPubKeysProvided
	}
	pubKey := pubKeys[0]
	headers := proof.GetHeaders(pubKey)
	if len(headers) == 0 {
		return nil, process.ErrNotEnoughHeadersProvided
	}

	return &proofTxData{
		round:     headers[0].Header.GetRound(),
		shardID:   headers[0].Header.GetShardID(),
		bytes:     proofBytes,
		slashType: proof.GetType(),
	}, nil
}
