package capnp

// AUTO GENERATED - DO NOT EDIT

import (
	"bufio"
	"bytes"
	"encoding/json"
	C "github.com/glycerine/go-capnproto"
	"io"
)

type LogLineMessageCapn C.Struct

func NewLogLineMessageCapn(s *C.Segment) LogLineMessageCapn {
	return LogLineMessageCapn(s.NewStruct(16, 2))
}
func NewRootLogLineMessageCapn(s *C.Segment) LogLineMessageCapn {
	return LogLineMessageCapn(s.NewRootStruct(16, 2))
}
func AutoNewLogLineMessageCapn(s *C.Segment) LogLineMessageCapn {
	return LogLineMessageCapn(s.NewStructAR(16, 2))
}
func ReadRootLogLineMessageCapn(s *C.Segment) LogLineMessageCapn {
	return LogLineMessageCapn(s.Root(0).ToStruct())
}
func (s LogLineMessageCapn) Message() string { return C.Struct(s).GetObject(0).ToText() }
func (s LogLineMessageCapn) MessageBytes() []byte {
	return C.Struct(s).GetObject(0).ToDataTrimLastByte()
}
func (s LogLineMessageCapn) SetMessage(v string)  { C.Struct(s).SetObject(0, s.Segment.NewText(v)) }
func (s LogLineMessageCapn) LogLevel() int32      { return int32(C.Struct(s).Get32(0)) }
func (s LogLineMessageCapn) SetLogLevel(v int32)  { C.Struct(s).Set32(0, uint32(v)) }
func (s LogLineMessageCapn) Args() C.TextList     { return C.TextList(C.Struct(s).GetObject(1)) }
func (s LogLineMessageCapn) SetArgs(v C.TextList) { C.Struct(s).SetObject(1, C.Object(v)) }
func (s LogLineMessageCapn) Timestamp() int64     { return int64(C.Struct(s).Get64(8)) }
func (s LogLineMessageCapn) SetTimestamp(v int64) { C.Struct(s).Set64(8, uint64(v)) }
func (s LogLineMessageCapn) WriteJSON(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('{')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"message\":")
	if err != nil {
		return err
	}
	{
		s := s.Message()
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
	_, err = b.WriteString("\"logLevel\":")
	if err != nil {
		return err
	}
	{
		s := s.LogLevel()
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
	_, err = b.WriteString("\"args\":")
	if err != nil {
		return err
	}
	{
		s := s.Args()
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
	err = b.WriteByte(',')
	if err != nil {
		return err
	}
	_, err = b.WriteString("\"timestamp\":")
	if err != nil {
		return err
	}
	{
		s := s.Timestamp()
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
func (s LogLineMessageCapn) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteJSON(&b)
	return b.Bytes(), err
}
func (s LogLineMessageCapn) WriteCapLit(w io.Writer) error {
	b := bufio.NewWriter(w)
	var err error
	var buf []byte
	_ = buf
	err = b.WriteByte('(')
	if err != nil {
		return err
	}
	_, err = b.WriteString("message = ")
	if err != nil {
		return err
	}
	{
		s := s.Message()
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
	_, err = b.WriteString("logLevel = ")
	if err != nil {
		return err
	}
	{
		s := s.LogLevel()
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
	_, err = b.WriteString("args = ")
	if err != nil {
		return err
	}
	{
		s := s.Args()
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
	_, err = b.WriteString(", ")
	if err != nil {
		return err
	}
	_, err = b.WriteString("timestamp = ")
	if err != nil {
		return err
	}
	{
		s := s.Timestamp()
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
func (s LogLineMessageCapn) MarshalCapLit() ([]byte, error) {
	b := bytes.Buffer{}
	err := s.WriteCapLit(&b)
	return b.Bytes(), err
}

type LogLineMessageCapn_List C.PointerList

func NewLogLineMessageCapnList(s *C.Segment, sz int) LogLineMessageCapn_List {
	return LogLineMessageCapn_List(s.NewCompositeList(16, 2, sz))
}
func (s LogLineMessageCapn_List) Len() int { return C.PointerList(s).Len() }
func (s LogLineMessageCapn_List) At(i int) LogLineMessageCapn {
	return LogLineMessageCapn(C.PointerList(s).At(i).ToStruct())
}
func (s LogLineMessageCapn_List) ToArray() []LogLineMessageCapn {
	n := s.Len()
	a := make([]LogLineMessageCapn, n)
	for i := 0; i < n; i++ {
		a[i] = s.At(i)
	}
	return a
}
func (s LogLineMessageCapn_List) Set(i int, item LogLineMessageCapn) {
	C.PointerList(s).Set(i, C.Object(item))
}
