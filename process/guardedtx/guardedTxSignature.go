package guardedtx

import (
	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/process/interceptors/processor"
)

// GuardianSigVerifier allows the verification of the guardian signatures for guarded transactions
// TODO: add an implementation and integrate it
type GuardianSigVerifier interface {
	VerifyGuardianSignature(account data.UserAccountHandler, inTx processor.InterceptedTransactionHandler) error
	IsInterfaceNil() bool
}

type guardedTxSigVerifier struct {
	sigVerifier     crypto.SingleSigner
	guardianChecker core.GuardianChecker
	encoder         data.Encoder
	marshaller      data.Marshaller
	keyGen          crypto.KeyGenerator
}

// NewGuardedTxSigVerifier creates a new instance of a guarded transaction signature verifier
func NewGuardedTxSigVerifier(sigVerifier crypto.SingleSigner, guardianChecker core.GuardianChecker) (*guardedTxSigVerifier, error) {
	if check.IfNil(sigVerifier) {
		return nil, process.ErrNilGuardianSigVerifier
	}
	if check.IfNil(guardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}

	return &guardedTxSigVerifier{
		sigVerifier:     sigVerifier,
		guardianChecker: guardianChecker,
	}, nil
}

// VerifyGuardianSignature verifies the guardian signature over the guarded transaction
func (gtx *guardedTxSigVerifier) VerifyGuardianSignature(account data.UserAccountHandler, inTx processor.InterceptedTransactionHandler) error {
	guardianPubKeyBytes, err := gtx.guardianChecker.GetActiveGuardian(account)
	if err != nil {
		return err
	}

	if len(guardianPubKeyBytes) == 0 {
		return process.ErrNilGuardianPublicKey
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

	guardianPubKey, err := gtx.keyGen.PublicKeyFromByteArray(guardianPubKeyBytes)
	if err != nil {
		return err
	}

	return gtx.sigVerifier.Verify(guardianPubKey, msgForSigVerification, guardedTxHandler.GetGuardianSignature())
}

// IsInterfaceNil returns nil if the receiver is nil
func (gtx *guardedTxSigVerifier) IsInterfaceNil() bool {
	return gtx == nil
}
