package proto

import "bytes"

type Close struct {}

func (*Close) Frontend() {}

func (msg *Close) Decode(src []byte) error {
	return nil
}

func (msg *Close) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('C')
	buf := new(bytes.Buffer)
	buf.WriteByte('P')
	buf.WriteString("")
	buf.WriteByte(0)
	return FinalizeMessage(dst, buf)
}