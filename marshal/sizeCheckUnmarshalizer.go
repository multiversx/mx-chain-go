package marshal

type sizeCheckUnmarshalizer struct {
	Marshalizer
	acceptedDelta uint32
}

// NewSizeCheckUnmarshalizer creates a wrapper around an existing marshalizer m
// which, during unmarshaling, also checks that the provided buffer dose not contain
// additional unused data.
func NewSizeCheckUnmarshalizer(m Marshalizer, maxDelta uint32) Marshalizer {
	scu := &sizeCheckUnmarshalizer{
		Marshalizer:   m,
		acceptedDelta: maxDelta,
	}
	return scu
}

// Unmarshal tries to unmarshal input buffer values into output object, and checks
// for additional unused data
func (scu *sizeCheckUnmarshalizer) Unmarshal(obj interface{}, buff []byte) error {
	err := scu.Marshalizer.Unmarshal(obj, buff)
	if err != nil {
		return err
	}

	out, erro := scu.Marshal(obj)
	if erro != nil {
		return err
	}

	maxSize := len(out) + len(out)*int(scu.acceptedDelta)/100
	if len(buff) > maxSize {
		return ErrUnmarshallingBadSize
	}
	return nil
}

// IsInterfaceNil returns true if there is no value under the interface or
// target marshalizer
func (scu *sizeCheckUnmarshalizer) IsInterfaceNil() bool {
	if scu != nil {
		return scu.Marshalizer == nil || scu.Marshalizer.IsInterfaceNil()
	}
	return true
}
