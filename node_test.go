package njson_test

import (
	"math"
	"testing"

	"github.com/alxarch/njson"
)

func TestNodeToFloat(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	n, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 1.2 {
		t.Errorf("Unexpected conversion %f", f)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 1.2 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if !math.IsNaN(f) {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToFloat(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %f", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if n, err := d.Parse("true"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToFloat(); ok {
		t.Errorf("Unexpected conversion error")
	}
}

func TestNodeToInt(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	n, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToInt(); ok {
		t.Errorf("Unexpected conversion error")
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != -17 {
		t.Errorf("Unexpected conversion %d", f)
	}
	if n, err := d.Parse("42.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	} else if f, ok := n.ToInt(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestNodeToUint(t *testing.T) {
	d := njson.BlankDocument()
	defer d.Close()
	n, err := d.Parse("1.2")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, err := d.Parse("NaN"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion error")
	}

	if n, err := d.Parse("0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 0 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-17"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if _, ok := n.ToUint(); ok {
		t.Errorf("Unexpected conversion ok")
	}

	if n, err := d.Parse("42.0"); err != nil {
		t.Errorf("Unexpected error: %s", err)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	} else if f, ok := n.ToUint(); !ok {
		t.Errorf("Unexpected conversion error")
	} else if f != 42 {
		t.Errorf("Unexpected conversion %d", f)
	}

	if n, err := d.Parse("-a7"); err == nil || n != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}
