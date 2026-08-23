package proto

import "bytes"

type CommandComplete struct {
	CommandTag []byte
}

func (*CommandComplete) Backend() {}

func (msg *CommandComplete) Decode(src []byte) error {
	idx := bytes.IndexByte(src, 0)
	msg.CommandTag = src[:idx]
	return nil
}

// Encode encodes src into dst. dst will include the 1 byte message type identifier and the 4 byte message length.
func (msg *CommandComplete) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}