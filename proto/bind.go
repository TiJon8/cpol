package proto

import (
	"bytes"
	"encoding/binary"

	"github.com/TiJon8/cpol/utils"
)

type Bind struct {
	Params []utils.DriverValue
}

func (*Bind) Frontend() {}

func (msg *Bind) Decode(src []byte) error {
	return nil
}

func (msg *Bind) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	dst.WriteByte('B')
	bufBind := new(bytes.Buffer)
	bufBind.WriteString("")
	bufBind.WriteByte(0)
	bufBind.WriteString("")
	bufBind.WriteByte(0)
	binary.Write(bufBind, binary.BigEndian, uint16(1))
	binary.Write(bufBind, binary.BigEndian, uint16(0))
	binary.Write(bufBind, binary.BigEndian, uint16(len(msg.Params)))
	for _, p := range msg.Params {
		if p.IsNull {
			binary.Write(bufBind, binary.BigEndian, int32(-1))
			continue
		}
		binary.Write(bufBind, binary.BigEndian, uint32(len(p.Value)))
		bufBind.WriteString(p.Value)
	}
	binary.Write(bufBind, binary.BigEndian, uint16(1))
	binary.Write(bufBind, binary.BigEndian, uint16(0))
	return FinalizeMessage(dst, bufBind)
}