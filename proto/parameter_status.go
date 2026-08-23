package proto

import (
	"bytes"
)

type ParameterStatus struct {
	Name string
	Value string
}

func (*ParameterStatus) Backend() {}


func (msg *ParameterStatus) Decode(src []byte) error {
	buf := bytes.NewBuffer(src)
	n, err := buf.ReadBytes(0)
	if err != nil {
		return err
	}
	name := string(n[:len(n)-1])

	v, err := buf.ReadBytes(0)
	if err != nil {
		return err
	}
	value := string(v[:len(v)-1])
	*msg = ParameterStatus{Name: name, Value: value}
	return  nil
}

func (src *ParameterStatus) Encode(dst *bytes.Buffer) (*bytes.Buffer, error) {
	return nil, nil
}
