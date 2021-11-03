package notifier

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	"github.com/ElrondNetwork/elrond-go/process"
)

type proofTxDataExtractor struct {
	marshaller marshal.Marshalizer
}

// NewProofTxDataExtractor creates a new instance of type ProofTxDataExtractor
func NewProofTxDataExtractor(marshaller marshal.Marshalizer) (ProofTxDataExtractor, error) {
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}

	return &proofTxDataExtractor{marshaller: marshaller}, nil
}

// GetProofTxData extracts proof tx data(see ProofTxData) from a slashing proof
func (ptd *proofTxDataExtractor) GetProofTxData(proof coreSlash.SlashingProofHandler) (*ProofTxData, error) {
	switch t := proof.(type) {
	case coreSlash.MultipleProposalProofHandler:
		return ptd.txDataFromMultipleHeaderProposalProof(t)
	case coreSlash.MultipleSigningProofHandler:
		return ptd.txDataFromMultipleHeaderSigningProof(t)
	default:
		return nil, process.ErrInvalidProof
	}
}

func (ptd *proofTxDataExtractor) txDataFromMultipleHeaderProposalProof(proof coreSlash.MultipleProposalProofHandler) (*ProofTxData, error) {
	proofBytes, err := ptd.marshaller.Marshal(proof)
	if err != nil {
		return nil, err
	}
	headers := proof.GetHeaders()
	if len(headers) == 0 {
		return nil, process.ErrNotEnoughHeadersProvided
	}

	return &ProofTxData{
		Round:     headers[0].GetHeaderHandler().GetRound(),
		ShardID:   headers[0].GetHeaderHandler().GetShardID(),
		Bytes:     proofBytes,
		SlashType: proof.GetType(),
	}, nil
}

func (ptd *proofTxDataExtractor) txDataFromMultipleHeaderSigningProof(proof coreSlash.MultipleSigningProofHandler) (*ProofTxData, error) {
	proofBytes, err := ptd.marshaller.Marshal(proof)
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
	if headers[0] == nil || check.IfNil(headers[0].GetHeaderHandler()) {
		return nil, process.ErrNilHeaderHandler
	}

	return &ProofTxData{
		Round:     headers[0].GetHeaderHandler().GetRound(),
		ShardID:   headers[0].GetHeaderHandler().GetShardID(),
		Bytes:     proofBytes,
		SlashType: proof.GetType(),
	}, nil
}

// IsInterfaceNil checks if the underlying pointer receiver is nil
func (ptd *proofTxDataExtractor) IsInterfaceNil() bool {
	return ptd == nil
}
