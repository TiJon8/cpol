package proto

import (
	"bytes"
	"encoding/binary"
)


type Terminate struct {}


func (*Terminate) Fronend() {}

func (msg *Terminate) Decode(src []byte) error {
	return nil
}

func (msg *Terminate) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('X')
	binary.Write(dst, binary.BigEndian, uint32(4))
	return dst, nil
}