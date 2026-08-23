package proto

import "bytes"

type Describe struct {}

func (*Describe) Frontend() {}

func (msg *Describe) Decode(src []byte) error {
	return nil
}

func (msg *Describe) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('D')
	buf := new(bytes.Buffer)
	buf.WriteByte('P')
	buf.WriteString("")
	buf.WriteByte(0)
	return FinalizeMessage(dst, buf)
}