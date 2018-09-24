package data

type Block struct {
	Nonce     int
	TimeStamp string
	Signature string
	Hash      string
	PrevHash  string
	MetaData  string
}

func NewBlock(nonce int, timeStamp string, signature string, hash string, prevhash string, metaData string) Block {

	b := Block{nonce, timeStamp, signature, hash, prevhash, metaData}
	return b
}

func (b *Block) SetNonce(nonce int) {
	b.Nonce = nonce
}

func (b *Block) GetNonce() int {
	return b.Nonce
}

func (b *Block) SetTimeStamp(timeStamp string) {
	b.TimeStamp = timeStamp
}

func (b *Block) GetTimeStamp() string {
	return b.TimeStamp
}

func (b *Block) SetSignature(signature string) {
	b.Signature = signature
}

func (b *Block) GetSignature() string {
	return b.Signature
}

func (b *Block) SetHash(hash string) {
	b.Hash = hash
}

func (b *Block) GetHash() string {
	return b.Hash
}

func (b *Block) SetPrevHash(prevHash string) {
	b.PrevHash = prevHash
}

func (b *Block) GetPrevHash() string {
	return b.PrevHash
}

func (b *Block) SetMetaData(metaData string) {
	b.MetaData = metaData
}

func (b *Block) GetMetaData() string {
	return b.MetaData
}
