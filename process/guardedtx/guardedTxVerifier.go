package guardedtx

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go-core/data"
	"github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/process"
	"github.com/ElrondNetwork/elrond-go/state"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
)

// GuardedTxSigVerifierArgs holds the argument to instantiate a guarded tx signature verifier
type GuardedTxSigVerifierArgs struct {
	SigVerifier     crypto.SingleSigner
	GuardianChecker process.GuardianChecker
	PubKeyConverter core.PubkeyConverter
	Marshaller      data.Marshaller
	KeyGen          crypto.KeyGenerator
}

type guardedTxSigVerifier struct {
	sigVerifier     crypto.SingleSigner
	guardianChecker process.GuardianChecker
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
func (gtx *guardedTxSigVerifier) VerifyGuardianSignature(account vmcommon.UserAccountHandler, inTx process.InterceptedTransactionHandler) error {
	guardianPubKey, err := gtx.GetGuardianPublicKey(account)
	if err != nil {
		return err
	}

	txHandler := inTx.Transaction()
	if check.IfNil(txHandler) {
		return process.ErrNilTransaction
	}

	guardedTxHandler, ok := txHandler.(data.GuardedTransactionHandler)
	if !ok {
		return process.ErrWrongTypeAssertion
	}

	inSignedTx, ok := inTx.(process.InterceptedSignedTransactionHandler)
	if !ok {
		return fmt.Errorf("%w to InterceptedSignedTransactionHandler", process.ErrWrongTypeAssertion)
	}

	msgForSigVerification, err := inSignedTx.GetTxMessageForSignatureVerification()
	if err != nil {
		return err
	}

	return gtx.sigVerifier.Verify(guardianPubKey, msgForSigVerification, guardedTxHandler.GetGuardianSignature())
}

// GetGuardianPublicKey returns the guardian public key for the given account
func (gtx *guardedTxSigVerifier) GetGuardianPublicKey(account vmcommon.UserAccountHandler) (crypto.PublicKey, error) {
	guardianPubKeyBytes, err := gtx.guardianChecker.GetActiveGuardian(account)
	if err != nil {
		return nil, err
	}

	if len(guardianPubKeyBytes) == 0 {
		return nil, process.ErrNilGuardianPublicKey
	}

	return gtx.keyGen.PublicKeyFromByteArray(guardianPubKeyBytes)
}

// HasPendingGuardian true if the given account has a pending guardian set
func (gtx *guardedTxSigVerifier) HasPendingGuardian(uah state.UserAccountHandler) bool {
	return gtx.guardianChecker.HasPendingGuardian(uah)
}

// IsInterfaceNil returns nil if the receiver is nil
func (gtx *guardedTxSigVerifier) IsInterfaceNil() bool {
	return gtx == nil
}
