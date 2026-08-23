package proto

import (
	"bytes"
	"strings"
)

type ErrorResponse struct {
	Error string
}

func (*ErrorResponse) Backend() {}


func (msg *ErrorResponse) Decode(src []byte) error {
	errBody := bytes.NewReader(src)
	var sb strings.Builder
	for {
		code, err := errBody.ReadByte()
		if err != nil {
			break
		}
		sb.WriteByte(code)
		var value []byte
		for {
			b, err := errBody.ReadByte()
			if err != nil || b == 0 {
				break
			}
			value = append(value, b)
		}
		sb.WriteString(": ")
		sb.Write(value)
		sb.WriteString(";")
		// fmt.Printf("Code: %s, value: %s\n", string(code), string(value))
	}
	msg.Error = sb.String()
	
	return nil
}

func (msg *ErrorResponse) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}