package njson

import (
	"io"
	"sync"
)

// Node is a reference to a node in a JSON Document.
// It is a versioned reference to avoid document manipulation after reset.
type Node struct {
	id  uint
	rev uint
	doc *Document
}

func (n Node) ID() uint {
	return n.id
}

func (n Node) With(id uint) Node {
	n.id = id
	return n
}

func (n Node) Document() *Document {
	if n.doc != nil && n.doc.rev == n.rev {
		return n.doc
	}
	return nil
}

func (p Node) N() *N {
	return p.Document().N(p.id)
}

// Lookup finds a node by path
func (p Node) Lookup(path []string) (Node, bool) {
	if id, ok := p.Document().Lookup(p.id, path); ok {
		return p.With(id), true
	}
	return p.With(maxUint), false
}

func (p Node) ToInterface() (interface{}, bool) {
	return p.Document().ToInterface(p.id)
}

var bufferpool = &sync.Pool{
	New: func() interface{} {
		return make([]byte, 2048)
	},
}

// Appender is a Marshaler interface for buffer append workflows.
type Appender interface {
	AppendJSON([]byte) ([]byte, error)
}

// PrintJSON is a helper to write an Appender to an io.Writer
func PrintJSON(w io.Writer, a Appender) (n int, err error) {
	b := bufferpool.Get().([]byte)
	if b, err = a.AppendJSON(b[:0]); err == nil {
		n, err = w.Write(b)
	}
	bufferpool.Put(b)
	return
}

func (p Node) PrintJSON(w io.Writer) (n int, err error) {
	return PrintJSON(w, p)
}
