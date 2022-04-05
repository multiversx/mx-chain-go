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
type GuardianSigVerifier interface {
	VerifyGuardianSignature(account data.UserAccountHandler, inTx processor.InterceptedTransactionHandler) error
	IsInterfaceNil() bool
}

type guardedTxSigVerifier struct {
	sigVerifier     crypto.SingleSigner
	guardianChecker core.GuardianChecker
	encoder         core.PubkeyConverter
	marshaller      data.Marshaller
	keyGen          crypto.KeyGenerator
}

// NewGuardedTxSigVerifier creates a new instance of a guarded transaction signature verifier
func NewGuardedTxSigVerifier(
	sigVerifier crypto.SingleSigner,
	guardianChecker core.GuardianChecker,
	pubKeyConverter core.PubkeyConverter,
	marshaller data.Marshaller,
	keyGen crypto.KeyGenerator,
) (*guardedTxSigVerifier, error) {
	if check.IfNil(sigVerifier) {
		return nil, process.ErrNilSingleSigner
	}
	if check.IfNil(guardianChecker) {
		return nil, process.ErrNilGuardianChecker
	}
	if check.IfNil(pubKeyConverter) {
		return nil, process.ErrNilPubkeyConverter
	}
	if check.IfNil(marshaller) {
		return nil, process.ErrNilMarshalizer
	}
	if check.IfNil(keyGen) {
		return nil, process.ErrNilKeyGen
	}

	return &guardedTxSigVerifier{
		sigVerifier:     sigVerifier,
		guardianChecker: guardianChecker,
		encoder:         pubKeyConverter,
		marshaller:      marshaller,
		keyGen:          keyGen,
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
