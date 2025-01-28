package external

// ArgsCreateTransaction defines arguments for creating a transaction
type ArgsCreateTransaction struct {
	Nonce               uint64
	Value               string
	Receiver            string
	ReceiverUsername    []byte
	Sender              string
	SenderUsername      []byte
	GasPrice            uint64
	GasLimit            uint64
	DataField           []byte
	SignatureHex        string
	ChainID             string
	Version             uint32
	Options             uint32
	Guardian            string
	GuardianSigHex      string
	Relayer             string
	RelayerSignatureHex string
}
