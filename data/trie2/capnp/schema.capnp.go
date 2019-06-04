package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type BranchNodeCapn C.Struct

func NewBranchNodeCapn(s *C.Segment) BranchNodeCapn      { return BranchNodeCapn(s.NewStruct(0, 1)) }
func NewRootBranchNodeCapn(s *C.Segment) BranchNodeCapn  { return BranchNodeCapn(s.NewRootStruct(0, 1)) }
func AutoNewBranchNodeCapn(s *C.Segment) BranchNodeCapn  { return BranchNodeCapn(s.NewStructAR(0, 1)) }
func ReadRootBranchNodeCapn(s *C.Segment) BranchNodeCapn { return BranchNodeCapn(s.Root(0).ToStruct()) }
func (s BranchNodeCapn) EncodedChildren() C.DataList     { return C.DataList(C.Struct(s).GetObject(0)) }
func (s BranchNodeCapn) SetEncodedChildren(v C.DataList) { C.Struct(s).SetObject(0, C.Object(v)) }
func (s BranchNodeCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"encodedChildren\":")
	if err != nil {
		return err
	}
	{
		s := s.EncodedChildren()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
		if err != nil {
			return err
		}
	}
	err = b.WriteByte('}')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s BranchNodeCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s BranchNodeCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("encodedChildren = ")
	if err != nil {
		return err
	}
	{
		s := s.EncodedChildren()
		{
			err = b.WriteByte('[')
			if err != nil {
				return err
			}
			for i, s := range s.ToArray() {
				if i != 0 {
					_, err = b.WriteString(", ")
				}
				if err != nil {
					return err
				}
				buf, err = json.Marshal(s)
				if err != nil {
					return err
				}
				_, err = b.Write(buf)
				if err != nil {
					return err
				}
			}
			err = b.WriteByte(']')
		}
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(')')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s BranchNodeCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type BranchNodeCapn_List C.PointerList

func NewBranchNodeCapnList(s *C.Segment, sz int) BranchNodeCapn_List {
	return BranchNodeCapn_List(s.NewCompositeList(0, 1, sz))
}
func (s BranchNodeCapn_List) Len() int { return C.PointerList(s).Len() }
func (s BranchNodeCapn_List) At(i int) BranchNodeCapn {
	return BranchNodeCapn(C.PointerList(s).At(i).ToStruct())
}
func (s BranchNodeCapn_List) ToArray() []BranchNodeCapn {
	n := s.Len()
	a := make([]BranchNodeCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s BranchNodeCapn_List) Set(i int, item BranchNodeCapn) { C.PointerList(s).Set(i, C.Object(item)) }

type ExtensionNodeCapn C.Struct

func NewExtensionNodeCapn(s *C.Segment) ExtensionNodeCapn { return ExtensionNodeCapn(s.NewStruct(0, 2)) }
func NewRootExtensionNodeCapn(s *C.Segment) ExtensionNodeCapn {
	return ExtensionNodeCapn(s.NewRootStruct(0, 2))
}
func AutoNewExtensionNodeCapn(s *C.Segment) ExtensionNodeCapn {
	return ExtensionNodeCapn(s.NewStructAR(0, 2))
}
func ReadRootExtensionNodeCapn(s *C.Segment) ExtensionNodeCapn {
	return ExtensionNodeCapn(s.Root(0).ToStruct())
}
func (s ExtensionNodeCapn) Key() []byte              { return C.Struct(s).GetObject(0).ToData() }
func (s ExtensionNodeCapn) SetKey(v []byte)          { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s ExtensionNodeCapn) EncodedChild() []byte     { return C.Struct(s).GetObject(1).ToData() }
func (s ExtensionNodeCapn) SetEncodedChild(v []byte) { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s ExtensionNodeCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"key\":")
	if err != nil {
		return err
	}
	{
		s := s.Key()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"encodedChild\":")
	if err != nil {
		return err
	}
	{
		s := s.EncodedChild()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte('}')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s ExtensionNodeCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s ExtensionNodeCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("key = ")
	if err != nil {
		return err
	}
	{
		s := s.Key()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("encodedChild = ")
	if err != nil {
		return err
	}
	{
		s := s.EncodedChild()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(')')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s ExtensionNodeCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type ExtensionNodeCapn_List C.PointerList

func NewExtensionNodeCapnList(s *C.Segment, sz int) ExtensionNodeCapn_List {
	return ExtensionNodeCapn_List(s.NewCompositeList(0, 2, sz))
}
func (s ExtensionNodeCapn_List) Len() int { return C.PointerList(s).Len() }
func (s ExtensionNodeCapn_List) At(i int) ExtensionNodeCapn {
	return ExtensionNodeCapn(C.PointerList(s).At(i).ToStruct())
}
func (s ExtensionNodeCapn_List) ToArray() []ExtensionNodeCapn {
	n := s.Len()
	a := make([]ExtensionNodeCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s ExtensionNodeCapn_List) Set(i int, item ExtensionNodeCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}

type LeafNodeCapn C.Struct

func NewLeafNodeCapn(s *C.Segment) LeafNodeCapn      { return LeafNodeCapn(s.NewStruct(0, 2)) }
func NewRootLeafNodeCapn(s *C.Segment) LeafNodeCapn  { return LeafNodeCapn(s.NewRootStruct(0, 2)) }
func AutoNewLeafNodeCapn(s *C.Segment) LeafNodeCapn  { return LeafNodeCapn(s.NewStructAR(0, 2)) }
func ReadRootLeafNodeCapn(s *C.Segment) LeafNodeCapn { return LeafNodeCapn(s.Root(0).ToStruct()) }
func (s LeafNodeCapn) Key() []byte                   { return C.Struct(s).GetObject(0).ToData() }
func (s LeafNodeCapn) SetKey(v []byte)               { C.Struct(s).SetObject(0, s.Segment.NewData(v)) }
func (s LeafNodeCapn) Value() []byte                 { return C.Struct(s).GetObject(1).ToData() }
func (s LeafNodeCapn) SetValue(v []byte)             { C.Struct(s).SetObject(1, s.Segment.NewData(v)) }
func (s LeafNodeCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"key\":")
	if err != nil {
		return err
	}
	{
		s := s.Key()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"value\":")
	if err != nil {
		return err
	}
	{
		s := s.Value()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte('}')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s LeafNodeCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s LeafNodeCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("key = ")
	if err != nil {
		return err
	}
	{
		s := s.Key()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("value = ")
	if err != nil {
		return err
	}
	{
		s := s.Value()
		buf, err = json.Marshal(s)
		if err != nil {
			return err
		}
		_, err = b.Write(buf)
		if err != nil {
			return err
		}
	}
	err = b.WriteByte(')')
	if err != nil {
		return err
	}
	err = b.Flush()
	return err
}
func (s LeafNodeCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type LeafNodeCapn_List C.PointerList

func NewLeafNodeCapnList(s *C.Segment, sz int) LeafNodeCapn_List {
	return LeafNodeCapn_List(s.NewCompositeList(0, 2, sz))
}
func (s LeafNodeCapn_List) Len() int { return C.PointerList(s).Len() }
func (s LeafNodeCapn_List) At(i int) LeafNodeCapn {
	return LeafNodeCapn(C.PointerList(s).At(i).ToStruct())
}
func (s LeafNodeCapn_List) ToArray() []LeafNodeCapn {
	n := s.Len()
	a := make([]LeafNodeCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s LeafNodeCapn_List) Set(i int, item LeafNodeCapn) { C.PointerList(s).Set(i, C.Object(item)) }
