package serializer

import (
	"errors"
	"fmt"
)

// partsHolder holds data parts (e.g. raw contract call arguments, raw contract return values).
// It allows one to easily construct parts (thus functioning as a builder of parts).
// It also allows one to focus on a specific part to read from (thus functioning as a reader of parts: think of a pick-up head).
// Both functionalities (building and reading) are kept within this single abstraction, for convenience.
type partsHolder struct {
	parts            [][]byte
	focusedPartIndex uint32
}

// newPartsHolder creates a new partsHolder, which has the given parts.
// Focus is on the first part, if any, or "beyond the last part" otherwise.
func newPartsHolder(parts [][]byte) *partsHolder {
	return &partsHolder{
		parts:            parts,
		focusedPartIndex: 0,
	}
}

// newEmptyPartsHolder creates a new partsHolder, which has no parts.
// Parts are created by calling appendEmptyPart().
// Focus is "beyond the last part" (since there is no part).
func newEmptyPartsHolder() *partsHolder {
	return &partsHolder{
		parts:            [][]byte{},
		focusedPartIndex: 0,
	}
}

func (holder *partsHolder) getParts() [][]byte {
	return holder.parts
}

func (holder *partsHolder) getNumParts() uint32 {
	return uint32(len(holder.parts))
}

func (holder *partsHolder) getPart(index uint32) ([]byte, error) {
	if index >= holder.getNumParts() {
		return nil, fmt.Errorf("part index %d is out of range", index)
	}

	return holder.parts[index], nil
}

func (holder *partsHolder) appendToLastPart(data []byte) error {
	if !holder.hasAnyPart() {
		return errors.New("cannot write, since there is no part to write to")
	}

	holder.parts[len(holder.parts)-1] = append(holder.parts[len(holder.parts)-1], data...)
	return nil
}

func (holder *partsHolder) hasAnyPart() bool {
	return len(holder.parts) > 0
}

func (holder *partsHolder) appendEmptyPart() {
	holder.parts = append(holder.parts, []byte{})
}

// readWholeFocusedPart reads the whole focused part, if any. Otherwise, it returns an error.
func (holder *partsHolder) readWholeFocusedPart() ([]byte, error) {
	if holder.isFocusedBeyondLastPart() {
		return nil, fmt.Errorf("cannot wholly read part %d: unexpected end of data", holder.focusedPartIndex)
	}

	part, err := holder.getPart(uint32(holder.focusedPartIndex))
	if err != nil {
		return nil, err
	}

	return part, nil
}

// focusOnNextPart focuses on the next part, if any. Otherwise, it returns an error.
func (holder *partsHolder) focusOnNextPart() error {
	if holder.isFocusedBeyondLastPart() {
		return fmt.Errorf(
			"cannot focus on next part, since the focus is already beyond the last part; focused part index is %d",
			holder.focusedPartIndex,
		)
	}

	holder.focusedPartIndex++
	return nil
}

// isFocusedBeyondLastPart returns true if the focus is already beyond the last part.
func (holder *partsHolder) isFocusedBeyondLastPart() bool {
	return holder.focusedPartIndex >= holder.getNumParts()
}
