package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)


type AuthenticationOk struct {}

func (*AuthenticationOk) Backend() {}

func (dst *AuthenticationOk) Decode(src []byte) error {
	authType := binary.BigEndian.Uint32(src)

	if authType != AuthTypeOk {
		return errors.New("bad auth type")
	}

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *AuthenticationOk) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}