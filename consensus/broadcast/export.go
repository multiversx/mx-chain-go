package broadcast

import (
	"github.com/ElrondNetwork/elrond-go-core/marshal"
	crypto "github.com/ElrondNetwork/elrond-go-crypto"
	"github.com/ElrondNetwork/elrond-go/consensus"
)

// SignMessage will sign and return the given message
func (cm *commonMessenger) SignMessage(message *consensus.Message) ([]byte, error) {
	return cm.peerSignatureHandler.GetPeerSignature(cm.privateKey, message.OriginatorPid)
}

// ExtractMetaMiniBlocksAndTransactions -
func (cm *commonMessenger) ExtractMetaMiniBlocksAndTransactions(
	miniBlocks map[uint32][]byte,
	transactions map[string][][]byte,
) (map[uint32][]byte, map[string][][]byte) {
	return cm.extractMetaMiniBlocksAndTransactions(miniBlocks, transactions)
}

// NewCommonMessenger will return a new instance of a commonMessenger
func NewCommonMessenger(
	marshalizer marshal.Marshalizer,
	messenger consensus.P2PMessenger,
	privateKey crypto.PrivateKey,
	shardCoordinator consensus.ShardCoordinator,
	peerSigHandler crypto.PeerSignatureHandler,
) (*commonMessenger, error) {

	return &commonMessenger{
		marshalizer:          marshalizer,
		messenger:            messenger,
		privateKey:           privateKey,
		shardCoordinator:     shardCoordinator,
		peerSignatureHandler: peerSigHandler,
	}, nil
}
