package mock

// BelNevMock is used to mock belare neven multisignature scheme
type BelNevMock struct {
	SetMessageMock           func(msg []byte)
	SetSigBitmapMock         func(bmp []byte) error
	SetAggregatedSigMock     func(aggSig []byte) error
	VerifyMock               func() error
	CreateCommitmentMock     func() (commSecret []byte, commitment []byte, err error)
	CommitmentBitmapMock     func() []byte
	AddCommitmentMock        func(index uint16, value []byte) error
	AggregateCommitmentsMock func() ([]byte, error)
	SetAggCommitmentMock     func(aggCommitment []byte, bitmap []byte) error
	SignPartialMock          func() ([]byte, error)
	SigBitmapMock            func() []byte
	AddSignPartialMock       func(index uint16, sig []byte) error
	VerifyPartialMock        func(index uint16, sig []byte) error
	AggregateSigsMock        func() ([]byte, error)
}

// SetMessage sets the message to be signed
func (bnm *BelNevMock) SetMessage(msg []byte) {
	bnm.SetMessageMock(msg)
}

// SetSigBitmap sets the signers bitmap. Starting with index 0, each signer has 1 bit according to it's position in the
// signers list, set to 1 if signer's signature is used, 0 if not used
func (bnm *BelNevMock) SetSigBitmap(bmp []byte) error {
	return bnm.SetSigBitmapMock(bmp)
}

// SetAggregatedSig sets the aggregated signature according to the given byte array
func (bnm *BelNevMock) SetAggregatedSig(aggSig []byte) error {
	return bnm.SetAggregatedSigMock(aggSig)
}

// Verify returns nil if the aggregateed signature is verified for the given public keys
func (bnm *BelNevMock) Verify() error {
	return bnm.VerifyMock()
}

// CreateCommitment creates a secret commitment and the corresponding public commitment point
func (bnm *BelNevMock) CreateCommitment() (commSecret []byte, commitment []byte, err error) {
	return bnm.CreateCommitmentMock()
}

// CommitmentBitmap returns the bitmap with the set
func (bnm *BelNevMock) CommitmentBitmap() []byte {
	return bnm.CommitmentBitmapMock()
}

// AddCommitment adds a commitment to the list on the specified position
func (bnm *BelNevMock) AddCommitment(index uint16, value []byte) error {
	return bnm.AddCommitmentMock(index, value)
}

// AggregateCommitments aggregates the list of commitments
func (bnm *BelNevMock) AggregateCommitments() ([]byte, error) {
	return bnm.AggregateCommitmentsMock()
}

// SetAggCommitment sets the aggregated commitment for the marked signers in bitmap
func (bnm *BelNevMock) SetAggCommitment(aggCommitment []byte, bitmap []byte) error {
	return bnm.SetAggCommitmentMock(aggCommitment, bitmap)
}

// SignPartial creates a partial signature
func (bnm *BelNevMock) SignPartial() ([]byte, error) {
	return bnm.SignPartialMock()
}

// SigBitmap returns the bitmap for the set partial signatures
func (bnm *BelNevMock) SigBitmap(bmp []byte) error {
	return bnm.SetSigBitmapMock(bmp)
}

// AddSignPartial adds the partial signature of the signer with specified position
func (bnm *BelNevMock) AddSignPartial(index uint16, sig []byte) error {
	return bnm.AddSignPartialMock(index, sig)
}

// VerifyPartial verifies the partial signature of the signer with specified position
func (bnm *BelNevMock) VerifyPartial(index uint16, sig []byte) error {
	return bnm.VerifyPartialMock(index, sig)
}

// AggregateSigs aggregates all collected partial signatures
func (bnm *BelNevMock) AggregateSigs() ([]byte, error) {
	return bnm.AggregateSigsMock()
}
