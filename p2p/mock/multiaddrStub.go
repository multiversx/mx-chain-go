package mock

import "github.com/multiformats/go-multiaddr"

// MultiaddrStub -
type MultiaddrStub struct {
	MarshalJSONCalled      func() ([]byte, error)
	UnmarshalJSONCalled    func(bytes []byte) error
	MarshalTextCalled      func() (text []byte, err error)
	UnmarshalTextCalled    func(text []byte) error
	MarshalBinaryCalled    func() (data []byte, err error)
	UnmarshalBinaryCalled  func(data []byte) error
	EqualCalled            func(multiaddr multiaddr.Multiaddr) bool
	BytesCalled            func() []byte
	StringCalled           func() string
	ProtocolsCalled        func() []multiaddr.Protocol
	EncapsulateCalled      func(multiaddr multiaddr.Multiaddr) multiaddr.Multiaddr
	DecapsulateCalled      func(multiaddr multiaddr.Multiaddr) multiaddr.Multiaddr
	ValueForProtocolCalled func(code int) (string, error)
}

// MarshalJSON -
func (mas *MultiaddrStub) MarshalJSON() ([]byte, error) {
	if mas.MarshalJSONCalled != nil {
		return mas.MarshalJSONCalled()
	}

	return nil, nil
}

// UnmarshalJSON -
func (mas *MultiaddrStub) UnmarshalJSON(bytes []byte) error {
	if mas.UnmarshalJSONCalled != nil {
		return mas.UnmarshalJSONCalled(bytes)
	}

	return nil
}

// MarshalText -
func (mas *MultiaddrStub) MarshalText() (text []byte, err error) {
	if mas.MarshalTextCalled != nil {
		return mas.MarshalTextCalled()
	}

	return nil, err
}

// UnmarshalText -
func (mas *MultiaddrStub) UnmarshalText(text []byte) error {
	if mas.UnmarshalTextCalled != nil {
		return mas.UnmarshalTextCalled(text)
	}

	return nil
}

// MarshalBinary -
func (mas *MultiaddrStub) MarshalBinary() (data []byte, err error) {
	if mas.MarshalBinaryCalled != nil {
		return mas.MarshalBinaryCalled()
	}

	return nil, nil
}

// UnmarshalBinary -
func (mas *MultiaddrStub) UnmarshalBinary(data []byte) error {
	if mas.UnmarshalBinaryCalled != nil {
		return mas.UnmarshalBinaryCalled(data)
	}

	return nil
}

// Equal -
func (mas *MultiaddrStub) Equal(multiaddr multiaddr.Multiaddr) bool {
	if mas.EqualCalled != nil {
		return mas.EqualCalled(multiaddr)
	}

	return false
}

// Bytes -
func (mas *MultiaddrStub) Bytes() []byte {
	if mas.BytesCalled != nil {
		return mas.BytesCalled()
	}

	return nil
}

// String -
func (mas *MultiaddrStub) String() string {
	if mas.StringCalled != nil {
		return mas.StringCalled()
	}

	return ""
}

// Protocols -
func (mas *MultiaddrStub) Protocols() []multiaddr.Protocol {
	if mas.ProtocolsCalled != nil {
		return mas.ProtocolsCalled()
	}

	return nil
}

// Encapsulate -
func (mas *MultiaddrStub) Encapsulate(multiaddr multiaddr.Multiaddr) multiaddr.Multiaddr {
	if mas.EncapsulateCalled != nil {
		return mas.EncapsulateCalled(multiaddr)
	}

	return nil
}

// Decapsulate -
func (mas *MultiaddrStub) Decapsulate(multiaddr multiaddr.Multiaddr) multiaddr.Multiaddr {
	if mas.DecapsulateCalled != nil {
		return mas.DecapsulateCalled(multiaddr)
	}

	return nil
}

// ValueForProtocol -
func (mas *MultiaddrStub) ValueForProtocol(code int) (string, error) {
	if mas.ValueForProtocolCalled != nil {
		return mas.ValueForProtocolCalled(code)
	}

	return "", nil
}
