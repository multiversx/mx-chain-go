package marshaling

// CreateMarshalizer creates a marshalizer with the specified kind
func CreateMarshalizer(kind MarshalizerKind) Marshalizer {
	switch kind {
	case JSON:
		return &jsonMarshalizer{}
	case Gob:
		return &gobMarshalizer{}
	default:
		return &jsonMarshalizer{}
	}
}
