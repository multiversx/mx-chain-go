package notifier

import (
	"errors"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/block"
	coreSlash "github.com/ElrondNetwork/elrond-go-core/data/slash"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const CommitmentProofValue = 1 // astea o sa fie in config undeva

type SlashingNotifier struct {
	privateKey      crypto.PrivateKey
	publicKey       crypto.PublicKey
	pubKeyConverter core.PubkeyConverter
	signer          crypto.SingleSigner
	accountHandler  vmcommon.AccountHandler // delete this
	marshaller      marshal.Marshalizer
	hasher          hashing.Hasher
}

func NewSlashingNotifier(
	signer crypto.SingleSigner,
	accountHandler vmcommon.AccountHandler,
) (slash.SlashingNotifier, error) {
	if check.IfNil(signer) {
		return nil, crypto.ErrNilSingleSigner
	}
	if check.IfNil(accountHandler) {
		return nil, state.ErrNilAccountHandler
	}

	return &SlashingNotifier{
		signer: signer,
	}, nil
}

func (sn *SlashingNotifier) CreateShardSlashingTransaction(proof slash.SlashingProofHandler) (data.TransactionHandler, error) {
	switch t := proof.(type) {
	case slash.MultipleProposalProofHandler:
		sn.createMultipleProposalProofTx(t)
	case slash.MultipleSigningProofHandler:
	default:
		return nil, errors.New("something")
	}

	return &transaction.Transaction{}, nil
}

func (sn *SlashingNotifier) createMultipleProposalProofTx(proof slash.MultipleProposalProofHandler) data.TransactionHandler {
	tx := &transaction.Transaction{
		Nonce:   sn.accountHandler.GetNonce(),
		Value:   big.NewInt(CommitmentProofValue),
		RcvAddr: nil, //core.MetachainShardId, , leave it dummy
		SndAddr: sn.accountHandler.AddressBytes(),
		Data:    nil,
	}

	protoProof := sn.convertToProtoProof(proof)
	proofBytes, _ := protoProof.Marshal()

	id := slash.ProofIDs[slash.MultipleProposal]
	proofHash := sn.hasher.Compute(string(proofBytes))
	crc := proofHash[len(proofHash)-2:]
	signatureProof, _ := sn.signer.Sign(sn.privateKey, proofHash)

	tmpData := make([]byte, 0)
	tmpData = append(tmpData, id)
	tmpData = append(tmpData, crc...)
	tmpData = append(tmpData, signatureProof...)

	tx.Data = tmpData

	return tx
}

func (sn *SlashingNotifier) convertToProtoProof(proof slash.MultipleProposalProofHandler) *coreSlash.MultipleHeaderProposalProof {
	interceptedHeaders := proof.GetHeaders()
	headers := make([]*block.Header, 0, len(interceptedHeaders))

	for _, interceptedHeader := range interceptedHeaders {
		hdr, castOk := interceptedHeader.HeaderHandler().(*block.Header)
		if !castOk {
			return nil
		}

		headers = append(headers, hdr)
	}

	return &coreSlash.MultipleHeaderProposalProof{
		Level: coreSlash.ThreatLevel(proof.GetLevel()),
		Headers: &coreSlash.Headers{
			Headers: headers,
		},
	}
}

func (sn *SlashingNotifier) CreateMetaSlashingEscalatedTransaction(slash.SlashingProofHandler) data.TransactionHandler {
	return nil
}
