package proto

import (
	"bytes"
	"encoding/binary"
)

type Parse struct {
	Query string
}

func (*Parse) Frontend() {}

func (msg *Parse) Decode(src []byte) error {
	return nil
}

func (msg *Parse) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('P')
	bufQuey := new(bytes.Buffer)
	bufQuey.WriteString("")
	bufQuey.WriteByte(0)
	bufQuey.WriteString(msg.Query)
	bufQuey.WriteByte(0)
	binary.Write(bufQuey, binary.BigEndian, uint16(0))
	return FinalizeMessage(dst, bufQuey)
}