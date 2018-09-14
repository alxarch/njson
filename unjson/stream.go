package unjson

import (
	"bufio"
	"errors"
	"io"
	"strings"

	"github.com/alxarch/njson"
)

// LineDecoder decodes  from a newline delimited JSON stream. (http://ndjson.org/)
type LineDecoder struct {
	Decoder                // Use a specific type decoder
	r       *bufio.Reader  // underlying reader
	p       njson.Document // a local njson.Parser
}

// NewLineDecoder creates a new LineDecoder
func NewLineDecoder(r io.Reader) *LineDecoder {
	if r == nil {
		return nil
	}
	d := LineDecoder{}
	if r, ok := r.(*bufio.Reader); ok {
		d.r = r
	} else {
		d.r = bufio.NewReader(r)
	}
	return &d
}

// Decode decodes the next JSON line in the stream to x
func (d *LineDecoder) Decode(x interface{}) (err error) {
	s := strings.Builder{}
	for {
		line, isPrefix, err := d.r.ReadLine()
		if err != nil {
			return err
		}
		s.Write(line)
		if isPrefix {
			continue
		}
		n, tail, err := d.p.Parse(s.String())
		if err != nil {
			return err
		}
		if strings.TrimSpace(tail) != "" {
			return errors.New("Invalid line delimited JSON")
		}
		if d.Decoder == nil {
			return UnmarshalFromNode(n, x)
		}
		return d.Decoder.Decode(x, n.ID(), &d.p)
	}
}

// LineEncoder encodes to a newline delimited JSON stream. (http://ndjson.org/)
type LineEncoder struct {
	Encoder
	buffer []byte
	w      io.Writer
}

// NewLineEncoder creates a new LineEncoder
func NewLineEncoder(w io.Writer) *LineEncoder {
	if w == nil {
		return nil
	}
	e := LineEncoder{w: w}
	return &e
}

// Encode encodes a value to a new JSON line on the stream.
func (e *LineEncoder) Encode(x interface{}) (err error) {
	if e.Encoder == nil {
		e.buffer, err = MarshalTo(e.buffer[:0], x)
	} else {
		e.buffer, err = e.Encoder.Encode(e.buffer[:0], x)
	}
	if err != nil {
		return
	}
	e.buffer = append(e.buffer, '\n')
	_, err = e.w.Write(e.buffer)
	return
}
