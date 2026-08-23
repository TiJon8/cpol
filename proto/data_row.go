package proto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type DataRow struct {
	Length int32
	Value []byte

	Rows []DataRow
}

func (*DataRow) Backend() {}

func (msg *DataRow) Decode(src []byte) error {
	drBody := bytes.NewReader(src)
	var valueCount uint16
	binary.Read(drBody, binary.BigEndian, &valueCount)
	var dataRows []DataRow
	fmt.Println(valueCount)
	for i := 0; i < int(valueCount); i++ {
		
		var valueLen int32
		binary.Read(drBody, binary.BigEndian, &valueLen)
		// по все видимсоти здесь и начинаются маппинги типов
		// рациональней value держать как []byte но тут string для облегчения и наглядного примера
		var value []byte
		if valueLen == -1 {
			value = nil
		} else {
			v := make([]byte, valueLen)
			io.ReadFull(drBody, v)
			value = v
		}
		dr := DataRow{
			Length: valueLen,
			Value: value,
		}
		dataRows = append(dataRows, dr)
	}
	msg.Rows = dataRows
	
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (msg *DataRow) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}