package common

// GetProofResponse is a struct that stores the response of a GetProof request
type GetProofResponse struct {
	Proof    [][]byte
	Value    []byte
	RootHash string
}
