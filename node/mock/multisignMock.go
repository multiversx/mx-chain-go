package mock

type MultisignMock struct {
}

func (mm *MultisignMock) Reset(pubKeys []string, index uint16) error {
	panic("implement me")
}

func (mm *MultisignMock) SetMessage(msg []byte) {
	panic("implement me")
}

func (mm *MultisignMock) SetAggregatedSig([]byte) error {
	panic("implement me")
}

func (mm *MultisignMock) Verify(bitmap []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) CreateCommitment() (commSecret []byte, commitment []byte, err error) {
	panic("implement me")
}

func (mm *MultisignMock) AddCommitmentHash(index uint16, commHash []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) CommitmentHash(index uint16) ([]byte, error) {
	panic("implement me")
}

func (mm *MultisignMock) SetCommitmentSecret(commSecret []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) AddCommitment(index uint16, value []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) Commitment(index uint16) ([]byte, error) {
	panic("implement me")
}

func (mm *MultisignMock) AggregateCommitments(bitmap []byte) ([]byte, error) {
	panic("implement me")
}

func (mm *MultisignMock) SetAggCommitment(aggCommitment []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) SignPartial(bitmap []byte) ([]byte, error) {
	panic("implement me")
}

func (mm *MultisignMock) AddSignPartial(index uint16, sig []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) VerifyPartial(index uint16, sig []byte, bitmap []byte) error {
	panic("implement me")
}

func (mm *MultisignMock) AggregateSigs(bitmap []byte) ([]byte, error) {
	panic("implement me")
}
