package block

type block struct {
	nonce     int
	timeStamp string
	signature string
	hash      string
	prevHash  string
	metaData  string
}

func New(nonce int, timeStamp string, signature string, hash string, prevhash string, metaData string) block {

	b := block{nonce, timeStamp, signature, hash, prevhash, metaData}
	return b
}

func (b block) setNonce(nonce int) {
	b.nonce = nonce
}

func (b block) getNonce() int {
	return b.nonce
}

func (b block) setTimeStamp(timeStamp string) {
	b.timeStamp = timeStamp
}

func (b block) getTimeStamp() string {
	return b.timeStamp
}

func (b block) setSignature(signature string) {
	b.signature = signature
}

func (b block) getSignature() string {
	return b.signature
}

func (b block) setHash(hash string) {
	b.hash = hash
}

func (b block) getHash() string {
	return b.hash
}

func (b block) setPrevHash(prevHash string) {
	b.prevHash = prevHash
}

func (b block) getPrevHash() string {
	return b.prevHash
}

func (b block) setMetaData(metaData string) {
	b.metaData = metaData
}

func (b block) getMetaData() string {
	return b.metaData
}
