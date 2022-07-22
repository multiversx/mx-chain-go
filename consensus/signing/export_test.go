package signing

func (sg *signer) GetSignatureShareByIndex(index uint16) ([]byte, error) {
	sg.mutSigningData.Lock()
	defer sg.mutSigningData.Unlock()

	if int(index) >= len(sg.data.sigShares) {
		return nil, ErrIndexOutOfBounds
	}

	return sg.data.sigShares[index], nil
}
