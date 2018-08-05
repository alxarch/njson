package scan

type Tokenizer struct {
	n     int // num tokens
	last  int // last stack
	stack []int
	Iterator
	tail int // last token
}

func (t *Tokenizer) Write(p []byte) (n int, err error) {
	t.tokens, n, err = Tokenize(t.tokens, p)
	return
}

type Iterator struct {
	tokens  Tokens
	path    []byte
	indexes []int
	head    int // first token
}

type ValueInfo uint8

const (
	_                       = iota
	valueNegative ValueInfo = 1 << iota
	valueScientific
	valueDecimal
	valueEmpty
	valueOK
)

type Value struct {
	tokens []Token
	str    string
	f64    float64
	u64    uint64
	typ    byte
	info   ValueInfo
}

func (v *Value) Iterator() Iterator {
	return Iterator{
		tokens: v.tokens,
	}
}

func (i *Iterator) Reset(tokens Tokens) {
	*i = Iterator{
		tokens:  tokens,
		path:    i.path[:0],
		indexes: i.indexes[:0],
	}
}

func (i *Iterator) ReadObject() (key []byte, value Value) {
	return
}

func (t *Tokenizer) Reset() {
	*t = Tokenizer{
		stack: t.stack[:0],
		Iterator: Iterator{
			tokens:  t.Iterator.tokens[:0],
			path:    t.Iterator.path[:0],
			indexes: t.Iterator.indexes[:0],
		},
	}
}
func (v *Value) Tokens() Tokens {
	return v.tokens
}

func (i *Iterator) next() *Token {
	return &i.tokens[i.head]
}

func (i *Iterator) Next() *Token {
	if 0 <= i.head && i.head < len(i.tokens) {
		return i.next()
	}
	return nil
}

func (i *Iterator) Scan(x interface{}) error {
	panic("Not yet implemented")
	return nil
}
