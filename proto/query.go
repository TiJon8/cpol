package proto

import (
	"bytes"
	"encoding/binary"
)


type Query struct {
	String string
}

// Frontend identifies this message as sendable by a PostgreSQL frontend.
func (*Query) Frontend() {}


func (src *Query) Decode(data []byte) error {
	return nil
}


func (src *Query) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('Q')
	bufQuery := new(bytes.Buffer)
	bufQuery.WriteString(src.String)
	bufQuery.WriteByte(0)
	binary.Write(dst, binary.BigEndian, uint32(4+bufQuery.Len()))
	dst.Write(bufQuery.Bytes())
	return dst, nil
}