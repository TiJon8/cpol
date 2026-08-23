package proto

import (
	"bytes"
	"encoding/binary"
)

type Flush struct {}

func (*Flush) Frontend() {}

func (msg *Flush) Decode(src []byte) error {
	return nil
}

func (msg *Flush) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('H')
	binary.Write(dst, binary.BigEndian, uint32(4))
	return dst, nil
}