package shard

import (
	"encoding/binary"
)

type Shard struct {
	ID uint32
}

func (s *Shard) IdToByteArray() ([]byte) {
	byteArrID := make([]byte, 4)
	binary.LittleEndian.PutUint32(byteArrID, s.ID)
	return byteArrID
}


