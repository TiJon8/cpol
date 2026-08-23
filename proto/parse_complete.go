package proto

import "bytes"

type ParseComplete struct {}


func (*ParseComplete) Backend() {}

func (msg *ParseComplete) Decode(src []byte) error {
	return nil
}

func (msg *ParseComplete) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}