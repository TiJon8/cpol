package proto

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	ProtocolVersionLatest = 196610
)

const (
	maxMessageBodyLen = (0x3fffffff - 1)
)

const (
	AuthTypeOk                = 0
	AuthTypeCleartextPassword = 3
)


type Frontend struct {
	r 			 io.Reader
	w 			io.Writer
	wbuf        *bytes.Buffer

	encodeError error
	partialMsg bool

	messageType byte
	bodyLen  int
	authType uint32

	authenticationOk                AuthenticationOk
	authenticationCleartextPassword  AuthenticationCleartextPassword
	startupMessage 				   StartupMessage
	terminate 					   Terminate
	backendKeyData				  BackendKeyData
	readyForQuery 				  ReadyForQuery
	passwordMessage 			  PasswordMessage
	errorResponse 				   ErrorResponse
	parameterStatus ParameterStatus
	rowDescription RowDescription
	dataRow DataRow
	commandComplete CommandComplete
	emptyQueryResponse EmptyQueryResponse
	parseComplete ParseComplete
	bindComplete BindComplete
	closeComplete CloseComplete
}

type Message interface {
	// Decode is allowed and expected to retain a reference to data after
	// returning (unlike encoding.BinaryUnmarshaler).
	Decode(data []byte) error

	// Encode appends itself to dst and returns the new buffer.
	Encode(dst *bytes.Buffer) (*bytes.Buffer, error)
}


type FrontendMessage interface {
	Message
	// Frontend() // no-op method to distinguish frontend from backend methods
}

type BackendMessage interface {
	Message
	Backend() // no-op method to distinguish frontend from backend methods
}


func NewFrontend(r io.Reader, w io.Writer) *Frontend {
	return &Frontend{
		r: r,
		w: w,
		wbuf: new(bytes.Buffer)}
}

func (f *Frontend) Send(msg FrontendMessage) {
	newBuf, err := msg.Encode(f.wbuf)
	if err != nil {
		f.encodeError = err
		return
	}
	f.wbuf = newBuf
}

func (f *Frontend) Flush() error {
	if err := f.encodeError; err != nil {
		f.encodeError = nil
		f.wbuf.Reset()
		return err
	}
	fmt.Println("flush frontend buffer", f.wbuf.Bytes())
	if f.wbuf.Len() == 0 {
		return nil
	}

	_, err := f.w.Write(f.wbuf.Bytes())

	const maxLen = 1024
	if f.wbuf.Len() > maxLen {
		f.wbuf = bytes.NewBuffer(make([]byte, 0, maxLen))
	} else {
		f.wbuf.Reset()
	}

	if err != nil {
		return err
	}

	return nil
}

func (f *Frontend) Next(n int) ([]byte, error) {
	read := make([]byte, n)
	_, err := io.ReadFull(f.r, read)
	if err != nil {
		return nil, err
	}
	return read, nil
}

func (f *Frontend) Receive() (BackendMessage, error) {
	if !f.partialMsg {
		header, err := f.Next(5)
		if err != nil {
			fmt.Println("cr error", err)
			return nil, err
		}

		f.messageType = header[0]
		msgLength := int(int32(binary.BigEndian.Uint32(header[1:])))
		if msgLength < 4 {
			return nil, fmt.Errorf("invalid message length: %d", msgLength)
		}
		f.bodyLen = msgLength - 4
		// fmt.Println("msg length: ", msgLength)
		f.partialMsg = true
	}

	
	msgBody, err := f.Next(f.bodyLen)
	if err != nil {
		return nil, err
	}

	f.partialMsg = false
	fmt.Println("msg type string", string(f.messageType))
	var msg BackendMessage
	switch f.messageType {
	case '1':
		msg = &f.parseComplete
	case '2':
		msg = &f.bindComplete
	case '3':
		msg = &f.closeComplete
	case 'K':
		msg = &f.backendKeyData
	case 'S':
		msg = &f.parameterStatus
	case 'R':
		var err error
		msg, err = f.findAuthenticationMessageType(msgBody)
		if err != nil {
			return nil, err
		}
	case 'T':
		msg = &f.rowDescription
	case 'D':
		msg = &f.dataRow
	case 'C':
		msg = &f.commandComplete
	case 'Z':
		msg = &f.readyForQuery
	case 'I':
		msg = &f.emptyQueryResponse
	case 'E':
		msg = &f.errorResponse
	default:
		return nil, fmt.Errorf("unknown message type: %c", f.messageType)
	}

	err = msg.Decode(msgBody)
	if err != nil {
		return nil, err
	}

	return msg, err
}

func (f *Frontend) findAuthenticationMessageType(src []byte) (BackendMessage, error) {
	f.authType = binary.BigEndian.Uint32(src[:4])
	switch f.authType {
	case AuthTypeOk:
		return &f.authenticationOk, nil
	case AuthTypeCleartextPassword:
		return &f.authenticationCleartextPassword, nil
	default:
		return nil, fmt.Errorf("unknown authentication type: %d", f.authType)
	}
}


func FinalizeMessage(dst *bytes.Buffer, buf *bytes.Buffer) (*bytes.Buffer, error) {
	messageBodyLen := buf.Len()
	if messageBodyLen > maxMessageBodyLen {
		return nil, errors.New("message body too large")
	}
	binary.Write(dst, binary.BigEndian, uint32(messageBodyLen)+4)
	dst.Write(buf.Bytes())
	return dst, nil
}