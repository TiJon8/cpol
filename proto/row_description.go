package proto

import (
	"bytes"
	"encoding/binary"
)

type RowDescription struct {
	FieldName string
	OIDTable uint32
	AttrNumTable uint16
	OIDType uint32
	TypeLen int16
	AttrTypeMod int32
	FormatCode uint16

	Rows []RowDescription
}

func (*RowDescription) Backend() {}

func (msg *RowDescription) Decode(src []byte) error {
	rdBody := bytes.NewReader(src)
	var fieldCount uint16
	binary.Read(rdBody, binary.BigEndian, &fieldCount)
	var rdSlice []RowDescription
	for i := 0; i < int(fieldCount); i++ {
		var fieldName []byte
		for {
			c, err := rdBody.ReadByte()
			if c == 0 || err != nil {
				break
			}
			fieldName = append(fieldName, c)
		}

		var oIDTable uint32
		var attrNumTable uint16
		var oIDType uint32
		var typeLen int16
		var attrTypeMod int32
		var formatCode uint16

		binary.Read(rdBody, binary.BigEndian, &oIDTable)
		binary.Read(rdBody, binary.BigEndian, &attrNumTable)
		binary.Read(rdBody, binary.BigEndian, &oIDType)
		binary.Read(rdBody, binary.BigEndian, &typeLen)
		binary.Read(rdBody, binary.BigEndian, &attrTypeMod)
		binary.Read(rdBody, binary.BigEndian, &formatCode)

		rdCol := RowDescription{
			FieldName: string(fieldName),
			OIDTable: oIDTable,
			AttrNumTable: attrNumTable,
			OIDType: oIDType,
			TypeLen: typeLen,
			AttrTypeMod: attrTypeMod,
			FormatCode: formatCode,
		}
		rdSlice = append(rdSlice, rdCol)
	}
	msg.Rows = rdSlice
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (msg *RowDescription) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}