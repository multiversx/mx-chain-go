package multisig_test

import (
	"testing"
)

func TestNewBelNevMultisig_NilHasherShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_NilPrivKeyShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_NilPubKeysShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_NilKeyGenShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_InvalidOwnIndexShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_InvalidPubKeyInListShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_EmptyPubKeyInListShouldErr(t *testing.T) {

}

func TestNewBelNevMultisig_OK(t *testing.T) {

}

func TestBelNevSigner_ResetNilPubKeysShouldErr(t *testing.T) {

}

func TestBelNevSigner_ResetInvalidPubKeyInListShouldErr(t *testing.T) {

}

func TestBelNevSigner_ResetEmptyPubKeyInListShouldErr(t *testing.T) {

}

func TestBelNevSigner_ResetOK(t *testing.T) {

}

func TestBelNevSigner_SetMessage(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentHashInvalidIndexShouldErr(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentHashNilCommitmentHashShouldErr(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentHashOK(t *testing.T) {

}

func TestBelNevSigner_CommitmentHashIndexOutOfBoundsShouldErr(t *testing.T) {

}

func TestBelNevSigner_CommitmentHashNotSetIndexShouldErr(t *testing.T) {

}

func TestBelNevSigner_CommitmentHashOK(t *testing.T) {

}

func TestBelNevSigner_CreateCommitmentWrongScalarMarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateCommitmentWrongPointMarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateCommitmentOK(t *testing.T) {

}

func TestBelNevSigner_SetCommitmentSecretWrongScalarUnmarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_SetCommitmentSecretOK(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentNilCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentWrongPointUnmarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentIndexOutOfBoundsShouldErr(t *testing.T) {

}

func TestBelNevSigner_AddCommitmentOK(t *testing.T) {

}

func TestBelNevSigner_CommitmentIndexOutOfBoundsShouldErr(t *testing.T) {

}

func TestBelNevSigner_CommitmentNilCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_CommitmentWrongPointMarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_CommitmentOK(t *testing.T) {

}

func TestBelNevSigner_AggregateCommitmentsNilBitmapShouldErr(t *testing.T) {

}

func TestBelNevSigner_AggregateCommitmentsWrongBitmapShouldErr(t *testing.T) {

}

func TestBelNevSigner_AggregateCommitmentsNotSetCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_AggregateCommitmentsWrongMarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_AggregateCommitmentsOK(t *testing.T) {

}

func TestBelNevSigner_SetAggCommitmentNilAggCommShouldErr(t *testing.T) {

}

func TestBelNevSigner_SetAggCommitmentOK(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareNilBitmapShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareInvalidIndexShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareNotSetCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareNotSetMessageShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareWrongPointMarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareNotSetAggCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareWrongScalarUnmarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareNotSetCommSecretShouldErr(t *testing.T) {

}

func TestBelNevSigner_CreateSignatureShareOK(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareNilSigShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareNilBitmapShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareNotSelectedIndexShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareWrongScalarUnmarshalShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareNotSetCommitmentShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareInvalidSignatureShouldErr(t *testing.T) {

}

func TestBelNevSigner_VerifySignatureShareOK(t *testing.T) {

}

