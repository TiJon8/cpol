package proto

import (
	"bytes"
)


type EmptyQueryResponse struct {}

func (*EmptyQueryResponse) Backend() {}

func (msg *EmptyQueryResponse) Decode(src []byte) error {
	return nil
}

func (msg *EmptyQueryResponse) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}