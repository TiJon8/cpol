package proto

import (
	"bytes"
	"encoding/binary"
)

type Execute struct {
	Query string
}

func (*Execute) Frontend() {}

func (msg *Execute) Decode(src []byte) error {
	return nil
}

func (msg *Execute) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('E')
	bufExec := new(bytes.Buffer)
	bufExec.WriteString("")
	bufExec.WriteByte(0)
	binary.Write(bufExec, binary.BigEndian, uint32(0))
	return FinalizeMessage(dst, bufExec)
}