package guardedtx

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
)

// GuardianSigVerifier allows the verification of the guardian signatures for guarded transactions
type GuardianSigVerifier interface {
	VerifyGuardianSignature(account data.UserAccountHandler, inTx process.InterceptedTransactionHandler) error
	IsInterfaceNil() bool
}

// GuardedTxSigVerifierArgs holds the argument to instantiate a guarded tx signature verifier
type GuardedTxSigVerifierArgs struct {
	SigVerifier     crypto.SingleSigner
	GuardianChecker core.GuardianChecker
	PubKeyConverter core.PubkeyConverter
	Marshaller      data.Marshaller
	KeyGen          crypto.KeyGenerator
}

type guardedTxSigVerifier struct {
	sigVerifier     crypto.SingleSigner
	guardianChecker core.GuardianChecker
	encoder         core.PubkeyConverter
	marshaller      data.Marshaller
	keyGen          crypto.KeyGenerator
}

// NewGuardedTxSigVerifier creates a new instance of a guarded transaction signature verifier
func NewGuardedTxSigVerifier(args GuardedTxSigVerifierArgs) (*guardedTxSigVerifier, error) {
	if check.IfNil(args.SigVerifier) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(args.GuardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}
	if check.IfNil(args.PubKeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(args.Marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(args.KeyGen) {
		return nil, process.ErrNilKeyGen
	}

	return &guardedTxSigVerifier{
		sigVerifier:     args.SigVerifier,
		guardianChecker: args.GuardianChecker,
		encoder:         args.PubKeyConverter,
		marshaller:      args.Marshaller,
		keyGen:          args.KeyGen,
	}, nil
}

// VerifyGuardianSignature verifies the guardian signature over the guarded transaction
func (gtx *guardedTxSigVerifier) VerifyGuardianSignature(account data.UserAccountHandler, inTx process.InterceptedTransactionHandler) error {
	guardianPubKey, err := gtx.GetGuardianPublicKey(account)
	if err != nil {
		return err
	}

	txHandler := inTx.Transaction()
	if check.IfNil(txHandler) {
		return process.ErrNilTransaction
	}

	guardedTxHandler, castOK := txHandler.(data.GuardedTransactionHandler)
	if !castOK {
		return process.ErrWrongTypeAssertion
	}

	msgForSigVerification, err := guardedTxHandler.GetDataForSigning(gtx.encoder, gtx.marshaller)
	if err != nil {
		return err
	}

	return gtx.sigVerifier.Verify(guardianPubKey, msgForSigVerification, guardedTxHandler.GetGuardianSignature())
}

// GetGuardianPublicKey returns the guardian public key for the given account
func (gtx *guardedTxSigVerifier) GetGuardianPublicKey(account data.UserAccountHandler) (crypto.PublicKey, error) {
	guardianPubKeyBytes, err := gtx.guardianChecker.GetActiveGuardian(account)
	if err != nil {
		return nil, err
	}

	if len(guardianPubKeyBytes) == 0 {
		return nil, process.ErrNilGuardianPublicKey
	}

	return gtx.keyGen.PublicKeyFromByteArray(guardianPubKeyBytes)
}

// IsInterfaceNil returns nil if the receiver is nil
func (gtx *guardedTxSigVerifier) IsInterfaceNil() bool {
	return gtx == nil
}
