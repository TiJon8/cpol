package proto

import "bytes"


type CloseComplete struct {}

func (*CloseComplete) Backend() {}

func (msg *CloseComplete) Decode(src []byte) error {
	return nil
}

func (msg *CloseComplete) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}