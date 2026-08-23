package proto

import (
	"bytes"
	"encoding/binary"
)

type Sync struct {}


func (*Sync) Frontend() {}

func (msg *Sync) Decode(src []byte) error {
	return nil
}

func (msg *Sync) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('S')
	binary.Write(dst, binary.BigEndian, uint32(4))
	return dst, nil
}