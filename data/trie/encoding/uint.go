package encoding

// Uint marshals/unmarshals as a JSON string with 0x prefix.
// The zero value marshals as "0x0".
type Uint uint

// MarshalText implements encoding.TextMarshaler.
func (b Uint) MarshalText() ([]byte, error) {
	return Uint64(b).MarshalText()
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Uint) UnmarshalJSON(input []byte) error {
	if !IsString(input) {
		return ErrNonString(uintT)
	}
	return WrapTypeError(b.UnmarshalText(input[1:len(input)-1]), uintT)
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Uint) UnmarshalText(input []byte) error {
	var u64 Uint64
	err := u64.UnmarshalText(input)
	if u64 > Uint64(^uint(0)) || err == ErrUint64Range {
		return ErrUintRange
	} else if err != nil {
		return err
	}
	*b = Uint(u64)
	return nil
}

// String returns the hex encoding of b.
func (b Uint) String() string {
	return EncodeUint64(uint64(b))
}
