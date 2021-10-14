package notifier

import (
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-core/data/transaction"
	"github.com/ElrondNetwork/elrond-go-core/hashing"
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/slash"
	"github.com/ElrondNetwork/elrond-go/state"
	"github.com/ElrondNetwork/elrond-go/update"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

const CommitmentProofValue = 1 // astea o sa fie in config undeva

type proto interface {
	Marshal() ([]byte, error)
}

type SlashingNotifierArgs struct {
	PrivateKey      crypto.PrivateKey
	PublicKey       crypto.PublicKey
	PubKeyConverter core.PubkeyConverter
	Signer          crypto.SingleSigner
	AccountHandler  vmcommon.AccountHandler // delete this
	Marshaller      marshal.Marshalizer
	Hasher          hashing.Hasher
}

type SlashingNotifier struct {
	privateKey      crypto.PrivateKey
	publicKey       crypto.PublicKey
	pubKeyConverter core.PubkeyConverter
	signer          crypto.SingleSigner
	accountHandler  vmcommon.AccountHandler // delete this
	marshaller      marshal.Marshalizer
	hasher          hashing.Hasher
}

func NewSlashingNotifier(args *SlashingNotifierArgs) (slash.SlashingNotifier, error) {
	if check.IfNil(args.PrivateKey) {
		return nil, crypto.ErrNilPrivateKey
	}
	if check.IfNil(args.PublicKey) {
		return nil, crypto.ErrNilPublicKey
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, update.ErrNilPubKeyConverter
	}
	if check.IfNil(args.Signer) {
		return nil, crypto.ErrNilSingleSigner
	}
	if check.IfNil(args.AccountHandler) {
		return nil, state.ErrNilAccountHandler
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}

	return &SlashingNotifier{
		privateKey:      args.PrivateKey,
		publicKey:       args.PublicKey,
		pubKeyConverter: args.PubKeyConverter,
		signer:          args.Signer,
		accountHandler:  args.AccountHandler,
		marshaller:      args.Marshaller,
		hasher:          args.Hasher,
	}, nil
}

func (sn *SlashingNotifier) CreateShardSlashingTransaction(proof slash.SlashingProofHandler) (data.TransactionHandler, error) {
	protoProof, err := toProto(proof)
	if err != nil {
		return nil, err
	}

	proofBytes, err := protoProof.Marshal()
	if err != nil {
		return nil, err
	}

	return sn.createProofTx(proof.GetType(), proofBytes)
}

func toProto(proof slash.SlashingProofHandler) (proto, error) {
	switch t := proof.(type) {
	case slash.MultipleProposalProofHandler:
		return slash.ToProtoMultipleHeaderProposal(t)
	case slash.MultipleSigningProofHandler:
		return slash.ToProtoMultipleHeaderSign(t)
	default:
		return nil, process.ErrUnknownProof
	}
}

func (sn *SlashingNotifier) createProofTx(slashType slash.SlashingType, proofBytes []byte) (*transaction.Transaction, error) {
	tx := &transaction.Transaction{
		Nonce:   sn.accountHandler.GetNonce(),
		Value:   big.NewInt(CommitmentProofValue),
		RcvAddr: nil, //core.MetachainShardId, , leave it dummy
		SndAddr: sn.accountHandler.AddressBytes(),
	}

	txData, err := sn.computeTxData(slashType, proofBytes)
	if err != nil {
		return nil, err
	}
	tx.Data = txData

	err = sn.signTx(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (sn *SlashingNotifier) computeTxData(slashType slash.SlashingType, proofBytes []byte) ([]byte, error) {
	proofHash := sn.hasher.Compute(string(proofBytes))
	proofSignature, err := sn.signer.Sign(sn.privateKey, proofHash)
	if err != nil {
		return nil, err
	}

	id, found := slash.ProofIDs[slashType]
	if !found {
		return nil, process.ErrUnknownProof
	}

	crc := proofHash[len(proofHash)-2:]

	dataStr := fmt.Sprintf("%s@%s@%s", []byte{id}, crc, proofSignature)
	return []byte(dataStr), nil
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

func (sn *SlashingNotifier) CreateMetaSlashingEscalatedTransaction(slash.SlashingProofHandler) data.TransactionHandler {
	return nil
}
