package unjson

import "github.com/alxarch/njson"

type RawString string

func (raw RawString) String() string {
	return string(raw)
}
func (raw *RawString) UnmarshalFromNode(n njson.Node) error {
	*raw = RawString(n.Raw())
	return nil
}

func (raw RawString) AppendJSON(dst []byte) ([]byte, error) {
	dst = append(dst, delimString)
	dst = append(dst, raw...)
	dst = append(dst, delimString)
	return dst, nil
}
