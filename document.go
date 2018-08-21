package njson

import (
	"math"
	"sync"
)

// Document is a json syntax tree.
type Document struct {
	nodes []Node   // All node offsets are indexes to this slice
	stack []uint16 // Parser stack
	n     uint16   // len(nodes)

	noCopy // protect from passing by value
}

// MaxDocumentSize is the maximum number of nodes a document can hold
const MaxDocumentSize = math.MaxUint16

// CopyTo all document nodes to another without allocation.
func (d *Document) CopyTo(c *Document) {
	*c = Document{
		nodes: append(c.nodes[:0], d.nodes...),
		n:     d.n,
	}
}

// Copy creates a copy of a document.
func (d *Document) Copy() *Document {
	c := Document{
		nodes: make([]Node, d.n),
		n:     d.n,
	}
	copy(c.nodes, d.nodes)
	return &c
}

// Reset resets a document to empty.
// This invalidates any Node pointers taken from this document.
func (d *Document) Reset() {
	d.nodes = d.nodes[:0]
	d.n = 0
	d.stack = d.stack[:0]
}

// add adds a Node for Token returning the new node's id
func (d *Document) add(t Token) (id uint16) {
	// Be safe and avoid adding of unescape token that doesn't exist
	t.extra = 0
	if id = d.n; id < MaxDocumentSize {
		d.nodes = append(d.nodes, Node{
			doc:    d,
			id:     d.n,
			parent: MaxDocumentSize,
			Token:  t,
		})
		d.n++
	}
	return

}

// Get finds a Node by id.
func (d *Document) Get(id uint16) *Node {
	if 0 <= id && id < d.n {
		return &d.nodes[id]
	}
	return nil
}

type noCopy struct{}

func (noCopy) Lock()   {}
func (noCopy) Unlock() {}

var docPool = &sync.Pool{
	New: func() interface{} {
		return new(Document)
	},
}

// BlankDocument returns a blank document from a pool.
// Put it back once you're done with Document.Close()
func BlankDocument() *Document {
	return docPool.Get().(*Document)
}

// Close returns the document to the pool.
func (d *Document) Close() error {
	if d == nil {
		return errNilDocument
	}
	d.Reset()
	docPool.Put(d)
	return nil
}
