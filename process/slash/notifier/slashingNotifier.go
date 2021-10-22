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
)

//TODO: Move this constants to config file, when slashing notifier is integrated.
// Please note that these are just some dummy values and are meant to be changed.

// BuiltInFunctionSlashCommitmentProof = key for slashing commitment proof built-in function
const BuiltInFunctionSlashCommitmentProof = "SlashCommitment"

// CommitmentProofValue = value to issue a commitment tx proof
const CommitmentProofValue = 1

// CommitmentProofGasPrice = gas price to issue a commitment tx proof
const CommitmentProofGasPrice = 1000000000

// CommitmentProofGasLimit = gas limit to issue a commitment tx proof
const CommitmentProofGasLimit = 70000

type proto interface {
	Marshal() ([]byte, error)
}

// SlashingNotifierArgs is a struct containing all arguments required to create a new slash.SlashingNotifier
type SlashingNotifierArgs struct {
	PrivateKey      crypto.PrivateKey
	PublicKey       crypto.PublicKey
	PubKeyConverter core.PubkeyConverter
	Signer          crypto.SingleSigner
	AccountAdapter  state.AccountsAdapter
	Hasher          hashing.Hasher
	Marshaller      marshal.Marshalizer
}

type slashingNotifier struct {
	privateKey      crypto.PrivateKey
	publicKey       crypto.PublicKey
	pubKeyConverter core.PubkeyConverter
	signer          crypto.SingleSigner
	accountAdapter  state.AccountsAdapter
	hasher          hashing.Hasher
	marshaller      marshal.Marshalizer
}

// NewSlashingNotifier creates a new instance of a slash.SlashingNotifier
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
	if check.IfNil(args.AccountAdapter) {
		return nil, state.ErrNilAccountsAdapter
	}
	if check.IfNil(args.Hasher) {
		return nil, process.ErrNilHasher
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}

	return &slashingNotifier{
		privateKey:      args.PrivateKey,
		publicKey:       args.PublicKey,
		pubKeyConverter: args.PubKeyConverter,
		signer:          args.Signer,
		accountAdapter:  args.AccountAdapter,
		hasher:          args.Hasher,
		marshaller:      args.Marshaller,
	}, nil
}

// CreateShardSlashingTransaction creates a so-called "commitment" transaction. If a slashing event has been detected,
// then a transaction will be issued, but it will not unveil details about the slash event, only a commitment proof.
// This tx is distinguished by its data field, which should be of format: ProofID@CRC@Sign(proof), where:
// 1. ProofID = 1 byte representing the slashing event ID (e.g.: multiple sign/proposal)
// 2. CRC = last 2 bytes of Hash(proof)
// 3. Sign(proof) = detector's proof signature. This is used to avoid front-running.
func (sn *slashingNotifier) CreateShardSlashingTransaction(proof slash.SlashingProofHandler) (data.TransactionHandler, error) {
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
		return nil, process.ErrInvalidProof
	}
}

func (sn *slashingNotifier) createProofTx(slashType slash.SlashingType, proofBytes []byte) (*transaction.Transaction, error) {
	tx, err := sn.createUnsignedTx(slashType, proofBytes)
	if err != nil {
		return nil, err
	}

	err = sn.signTx(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (sn *slashingNotifier) createUnsignedTx(slashType slash.SlashingType, proofBytes []byte) (*transaction.Transaction, error) {
	pubKey, err := sn.publicKey.ToByteArray()
	if err != nil {
		return nil, err
	}
	account, err := sn.accountAdapter.GetExistingAccount(pubKey)
	if err != nil {
		return nil, err
	}
	txData, err := sn.computeTxData(slashType, proofBytes)
	if err != nil {
		return nil, err
	}

	return &transaction.Transaction{
		Nonce:    account.GetNonce(),
		Value:    big.NewInt(CommitmentProofValue),
		RcvAddr:  nil, //TODO: This should be changed to a meta chain address
		SndAddr:  account.AddressBytes(),
		GasPrice: CommitmentProofGasPrice,
		GasLimit: CommitmentProofGasLimit,
		Data:     txData,
	}, nil
}

func (sn *slashingNotifier) computeTxData(slashType slash.SlashingType, proofBytes []byte) ([]byte, error) {
	proofHash := sn.hasher.Compute(string(proofBytes))
	proofSignature, err := sn.signer.Sign(sn.privateKey, proofHash)
	if err != nil {
		return nil, err
	}

	id, found := slash.ProofIDs[slashType]
	if !found {
		return nil, process.ErrInvalidProof
	}

	crc := proofHash[len(proofHash)-2:]

	dataStr := fmt.Sprintf("%s@%s@%s@%s", BuiltInFunctionSlashCommitmentProof, []byte{id}, crc, proofSignature)
	return []byte(dataStr), nil
}

func (sn *slashingNotifier) signTx(tx *transaction.Transaction) error {
	txBytes, err := tx.GetDataForSigning(sn.pubKeyConverter, sn.marshaller)
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

// CreateMetaSlashingEscalatedTransaction currently not implemented
func (sn *slashingNotifier) CreateMetaSlashingEscalatedTransaction(slash.SlashingProofHandler) data.TransactionHandler {
	return nil
}
