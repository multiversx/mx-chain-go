package notifier

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
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
	var tx *transaction.Transaction
	var err error

	switch proofType := proof.(type) {
	case slash.MultipleProposalProofHandler:
		tx, err = sn.createMultipleProposalProofTx(proofType)
	case slash.MultipleSigningProofHandler:
		tx, err = sn.createMultipleSignProofTx(proofType)
	default:
		return nil, errors.New("something")
	}

	if err != nil {
		return nil, err
	}

	err = sn.signTx(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (sn *SlashingNotifier) createMultipleProposalProofTx(proof slash.MultipleProposalProofHandler) (*transaction.Transaction, error) {
	tx := sn.createUnsignedTx()

	protoProof, err := slash.ToProtoMultipleHeaderProposal(proof)
	if err != nil {
		return nil, err
	}

	proofBytes, err := protoProof.Marshal()
	if err != nil {
		return nil, err
	}

	txData, err := sn.computeTxProofData(slash.MultipleProposal, proofBytes)
	if err != nil {
		return nil, err
	}

	tx.Data = txData
	return tx, nil
}

func (sn *SlashingNotifier) computeTxProofData(slashType slash.SlashingType, proofBytes []byte) ([]byte, error) {
	proofHash := sn.hasher.Compute(string(proofBytes))
	id := slash.ProofIDs[slashType]
	crc := proofHash[len(proofHash)-2:]
	signatureProof, err := sn.signer.Sign(sn.privateKey, proofHash)
	if err != nil {
		return nil, err
	}

	dataStr := fmt.Sprintf("%v@%v@%v", id, crc, signatureProof)
	return []byte(dataStr), nil
}

func (sn *SlashingNotifier) createUnsignedTx() *transaction.Transaction {
	return &transaction.Transaction{
		Nonce:   sn.accountHandler.GetNonce(),
		Value:   big.NewInt(CommitmentProofValue),
		RcvAddr: nil, //core.MetachainShardId, , leave it dummy
		SndAddr: sn.accountHandler.AddressBytes(),
		Data:    nil,
	}
}

func (sn *SlashingNotifier) createMultipleSignProofTx(proof slash.MultipleSigningProofHandler) (*transaction.Transaction, error) {
	tx := sn.createUnsignedTx()

	protoProof, err := slash.ToProtoMultipleHeaderSign(proof)
	if err != nil {
		return nil, err
	}

	proofBytes, err := protoProof.Marshal()
	if err != nil {
		return nil, err
	}

	txData, err := sn.computeTxProofData(slash.MultipleSigning, proofBytes)
	if err != nil {
		return nil, err
	}

	tx.Data = txData
	return tx, nil
}

func (sn *SlashingNotifier) CreateMetaSlashingEscalatedTransaction(slash.SlashingProofHandler) data.TransactionHandler {
	return nil
}

func (sn *SlashingNotifier) signTx(tx *transaction.Transaction) error {
	txBytes, err := tx.Marshal()
	if err != nil {
		return err
	}

	signature, err := sn.signer.Sign(sn.privateKey, txBytes)
	if err != nil {
		return err
	}

	tx.Signature = signature
	return nil
}
