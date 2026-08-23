package proto

import (
	"bytes"
	"encoding/binary"
)

type PasswordMessage struct {
	Password string
}

func (msg *PasswordMessage) Decode(src []byte) error {
	return nil
}

func (msg *PasswordMessage) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('p')

	bbody := new(bytes.Buffer)
	bbody.WriteString("root")
	bbody.WriteByte(0)

	binary.Write(dst, binary.BigEndian, uint32(bbody.Len()+4))
	dst.Write(bbody.Bytes())
	return dst, nil

	// fmt.Println(dst.Len())
	// dst.WriteByte('p')
	// buf := new(bytes.Buffer)
	// buf.WriteString(msg.Password)
	// buf.WriteByte(0)
	// return FinalizeMessage(dst, buf)
}