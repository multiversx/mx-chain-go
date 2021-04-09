package block

import "github.com/ElrondNetwork/elrond-go/data"

// SetType -
func (m *MiniBlockHeader) GetTypeInt32() int32 {
	if m == nil {
		return -1
	}
	return int32(m.Type)
}

// SetHash -
func (m *MiniBlockHeader) SetHash(hash []byte) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.Hash = hash
	return nil
}

// SetSenderShardID -
func (m *MiniBlockHeader) SetSenderShardID(shardID uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.SenderShardID = shardID
	return nil
}

// SetReceiverShardID -
func (m *MiniBlockHeader) SetReceiverShardID(shardID uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.ReceiverShardID = shardID
	return nil
}

// SetTxCount -
func (m *MiniBlockHeader) SetTxCount(count uint32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.TxCount = count
	return nil
}

// SetType -
func (m *MiniBlockHeader) SetType(t int32) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.Type = Type(t)
	return nil
}

// SetReserved -
func (m *MiniBlockHeader) SetReserved(reserved []byte) error {
	if m == nil {
		return data.ErrNilPointerReceiver
	}
	m.Reserved = reserved
	return nil
}

// ShallowClone -
func (m *MiniBlockHeader) ShallowClone() data.MiniBlockHeaderHandler {
	if m == nil {
		return nil
	}
	mbhCopy := *m
	return &mbhCopy
}
