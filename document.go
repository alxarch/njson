package njson

import "math"

type Document struct {
	nodes []Node
	n     uint16

	noCopy
}

func (d *Document) CopyTo(c *Document) {
	*c = Document{
		nodes: append(c.nodes[:0], d.nodes...),
		n:     d.n,
	}
}

func (d *Document) Copy() *Document {
	c := Document{
		nodes: make([]Node, d.n),
		n:     d.n,
	}
	copy(c.nodes, d.nodes)
	return &c
}

func (d *Document) Reset() {
	d.nodes = d.nodes[:0]
	d.n = 0
}

func (d *Document) add(t Token) (id uint16) {
	if id = d.n; id < math.MaxUint16 {
		d.nodes = append(d.nodes, Node{
			doc:   d,
			id:    d.n,
			Token: t,
		})
		d.n++
	}
	return

}

func (d *Document) CreateNode(src string) (*Node, error) {
	p := DocumentParser{}
	id, err := p.Parse(src, d)
	if err == nil {
		return &d.nodes[id], nil
	}
	d.nodes = d.nodes[:id]
	d.n = id
	return nil, err
}

func (d *Document) Get(id uint16) *Node {
	if 0 <= id && id < d.n {
		return &d.nodes[id]
	}
	return nil
}

func (d *Document) get(id uint16) *Node {
	if 0 < id && id < d.n {
		return &d.nodes[id]
	}
	return nil
}

type noCopy struct{}

func (noCopy) Lock()   {}
func (noCopy) Unlock() {}
