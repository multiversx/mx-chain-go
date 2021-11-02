package notifier

import (
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type proofTxData struct {
	round     uint64
	shardID   uint32
	bytes     []byte
	slashType coreSlash.SlashingType
}

func txDataFromMultipleHeaderProposalProof(marshaller marshal.Marshalizer, proof coreSlash.MultipleProposalProofHandler) (*proofTxData, error) {
	proofBytes, err := marshaller.Marshal(proof)
	if err != nil {
		return nil, err
	}
	headers := proof.GetHeaders()
	if len(headers) == 0 {
		return nil, process.ErrNotEnoughHeadersProvided
	}

	return &proofTxData{
		round:     headers[0].GetHeaderHandler().GetRound(),
		shardID:   headers[0].GetHeaderHandler().GetShardID(),
		bytes:     proofBytes,
		slashType: proof.GetType(),
	}, nil
}

func txDataFromMultipleHeaderSigningProof(marshaller marshal.Marshalizer, proof coreSlash.MultipleSigningProofHandler) (*proofTxData, error) {
	proofBytes, err := marshaller.Marshal(proof)
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
		round:     headers[0].GetHeaderHandler().GetRound(),
		shardID:   headers[0].GetHeaderHandler().GetShardID(),
		bytes:     proofBytes,
		slashType: proof.GetType(),
	}, nil
}
