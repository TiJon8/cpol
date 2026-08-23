package proto

import "bytes"

type BindComplete struct {}

func (*BindComplete) Backend() {}

func (msg *BindComplete) Decode(src []byte) error {
	return nil
}

func (msg *BindComplete) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}