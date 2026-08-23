package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
)


type AuthenticationCleartextPassword struct {}


func (*AuthenticationCleartextPassword) Backend() {}

func (dst *AuthenticationCleartextPassword) Decode(src []byte) error {
	authType := binary.BigEndian.Uint32(src)

	if authType != AuthTypeCleartextPassword {
		return errors.New("bad auth type")
	}

	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *AuthenticationCleartextPassword) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}