package proto

import (
	"bytes"
	"encoding/binary"
	"io"
)


type BackendKeyData struct {
	ProcessID uint32
	SecretKey []byte
}

func (BackendKeyData) Backend() {}

func (dst *BackendKeyData) Decode(src []byte) error {
	reader := bytes.NewReader(src)
	pids := make([]byte, 4)
	reader.Read(pids)
	dst.ProcessID = binary.BigEndian.Uint32(pids)
	dst.SecretKey, _ = io.ReadAll(reader)
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (src *BackendKeyData) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}