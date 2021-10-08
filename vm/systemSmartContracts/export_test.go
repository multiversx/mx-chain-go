package systemSmartContracts

//RemoveFromWaitingList -
func (s *stakingSC) RemoveFromWaitingList(blsKey []byte) error {
	return s.removeFromWaitingList(blsKey)
}
