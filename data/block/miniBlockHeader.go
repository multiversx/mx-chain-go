package block

import "github.com/ElrondNetwork/elrond-go/data"

// SetType -
func (m *MiniBlockHeader) GetTypeInt32() int32 {
	return int32(m.Type)
}

// SetHash -
func (m *MiniBlockHeader) SetHash(hash []byte) {
	m.Hash = hash
}

// SetSenderShardID -
func (m *MiniBlockHeader) SetSenderShardID(shardID uint32) {
	m.SenderShardID = shardID
}

// SetReceiverShardID -
func (m *MiniBlockHeader) SetReceiverShardID(shardID uint32) {
	m.ReceiverShardID = shardID
}

// SetTxCount -
func (m *MiniBlockHeader) SetTxCount(count uint32) {
	m.TxCount = count
}

// SetType -
func (m *MiniBlockHeader) SetType(t int32) {
	m.Type = Type(t)
}

// SetReserved -
func (m *MiniBlockHeader) SetReserved(reserved []byte) {
	m.Reserved = reserved
}

// ShallowClone -
func (m *MiniBlockHeader) ShallowClone() data.MiniBlockHeaderHandler {
	mbhCopy := *m
	return &mbhCopy
}
