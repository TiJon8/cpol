package proto

import (
	"bytes"
)


type ReadyForQuery struct {
	TxStatus byte
}

// Backend identifies this message as sendable by the PostgreSQL backend.
func (*ReadyForQuery) Backend() {}

// Decode decodes src into dst. src must contain the complete message with the exception of the initial 1 byte message
// type identifier and 4 byte message length.
func (dst *ReadyForQuery) Decode(src []byte) error {

	dst.TxStatus = src[0]
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *ReadyForQuery) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}