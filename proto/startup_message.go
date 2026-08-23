package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)




type StartupMessage struct {
	ProtocolVersion uint32
	Parameters      map[string]string
}

func (*StartupMessage) Frontend() {}


func (src *StartupMessage) Decode(data []byte) error {
	return nil
}


func (src *StartupMessage) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {

// 	dst = AppendUint32(dst, src.ProtocolVersion)
// 	fmt.Println(dst.String())
// 	dst.WriteByte(0)
// 	return finishMessage(dst)
	buf := new(bytes.Buffer)
	binary.Write(dst, binary.BigEndian, src.ProtocolVersion)
	// User specified
	for k, v := range src.Parameters {
		dst.WriteString(k)
		dst.WriteByte(0)
		dst.WriteString(v)
		dst.WriteByte(0)
	}
	// zero byte in the end of the list
	dst.WriteByte(0)
	binary.Write(buf, binary.BigEndian, uint32(dst.Len()+4))
	buf.Write(dst.Bytes())
	return buf, nil
}

func finishMessage(dst *bytes.Buffer) (*bytes.Buffer, error) {
	messageBodyLen := dst.Len()
	if messageBodyLen > maxMessageBodyLen {
		return nil, errors.New("message body too large")
	}
	finalize := new(bytes.Buffer)
	binary.Write(finalize, binary.BigEndian, uint32(messageBodyLen))
	finalize.Write(dst.Bytes())
	fmt.Println(finalize.String())
	return finalize, nil
}

// func AppendInt32(buf bytes.Buffer, n int32) []byte {
// 	binary.Write(b, binary.BigEndian, uint32(n))
// 	b.Write(b.Bytes())
// 	appended := AppendUint32(buf, uint32(n))
// }

func AppendUint32(buf *bytes.Buffer, n uint32) *bytes.Buffer {
	binary.Write(buf, binary.BigEndian, uint32(n))
	return buf
}
